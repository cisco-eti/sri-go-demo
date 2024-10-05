package logger

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

const uuidStringLength = 36
const uuidBytesLength = 16
const validYears = 10

// UUID as a []byte array
type UUID [uuidBytesLength]byte

// NewUUID return V1 uuid that has first 4 bytes as unix timestamp
func NewUUID() UUID {
	ud := UUID{}
	binary.BigEndian.PutUint32(ud[0:], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(ud[4:], rand.Uint32())
	binary.BigEndian.PutUint64(ud[8:], rand.Uint64())
	return ud
}

// Bytes return as byte array
func (ud UUID) Bytes() []byte {
	return ud[:]
}

// String return string
func (ud UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", ud[0:4], ud[4:6], ud[6:8], ud[8:10], ud[10:])
}

// GetTimeStamp return time from uuid
func (ud UUID) GetTimeStamp() (time.Time, error) {
	buf := bytes.NewReader(ud[0:4])
	var unix uint32
	if err := binary.Read(buf, binary.BigEndian, &unix); err != nil {
		return time.Time{}, err
	}
	stamp := time.Unix(int64(unix), 0)
	if !IsTimeStampValid(stamp) {
		return time.Time{}, fmt.Errorf("uuid not before now and after %d years ago stamp=%s", validYears, stamp.Format(time.RFC3339))
	}
	return stamp, nil
}

// IsTimeStampValid if time before now and after -validYears
func IsTimeStampValid(stamp time.Time) bool {
	now := time.Now().AddDate(0, 0, 1)
	if stamp.After(now) {
		return false
	}
	if stamp.Before(now.AddDate(-validYears, 0, 0)) {
		return false
	}
	return true
}

// MarshalJSON Json
func (ud UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(ud.String())
}

// MarshalText text
// The encoding is the same as returned by String.
func (ud UUID) MarshalText() ([]byte, error) {
	return []byte(ud.String()), nil
}

// MarshalBinary binary
func (ud UUID) MarshalBinary() ([]byte, error) {
	return ud.Bytes(), nil
}

// UnmarshalText text
func (ud *UUID) UnmarshalText(text []byte) error {
	byteGroups := []int{8, 4, 4, 4, 12}
	if len(text) != uuidStringLength {
		return fmt.Errorf("uuid: incorrect UUID length: %s", text)
	}
	src := text
	dst := ud[:]
	if src[8] != '-' || src[13] != '-' || src[18] != '-' || src[23] != '-' {
		return fmt.Errorf("uuid: incorrect UUID format %s", src)
	}
	for i, byteGroup := range byteGroups {
		if i > 0 {
			src = src[1:] // skip dash
		}
		_, err := hex.Decode(dst[:byteGroup/2], src[:byteGroup])
		if err != nil {
			return err
		}
		src = src[byteGroup:]
		dst = dst[byteGroup/2:]
	}
	return nil
}

// UnmarshalBinary binary
func (ud *UUID) UnmarshalBinary(data []byte) error {
	if len(data) != uuidBytesLength {
		return fmt.Errorf("uuid: UUID must be exactly 16 bytes long, got %d bytes", len(data))
	}
	copy(ud[:], data)
	return nil
}

// FromString get uuid from string
func FromString(input string) (UUID, error) {
	var ud UUID
	err := ud.UnmarshalText([]byte(input))
	return ud, err
}

// Scan implements the sql.Scanner interface
func (ud *UUID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return ud.UnmarshalBinary(src)

	case string:
		return ud.UnmarshalText([]byte(src))
	}
	return fmt.Errorf("uuid: cannot convert %T to UUID", src)
}

// MarshalJSON Json
func (nu NullUUID) MarshalJSON() ([]byte, error) {
	if !nu.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nu.UUID)
}

// NullUUID value that can be NULL in the database
type NullUUID struct {
	UUID  UUID
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (nu *NullUUID) Scan(src interface{}) error {
	if src == nil {
		nu.Valid = false
		return nil
	}
	// Delegate to UUID Scan function
	nu.Valid = true
	return nu.UUID.Scan(src)
}
