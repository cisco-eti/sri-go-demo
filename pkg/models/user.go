package models

import "gorm.io/gorm"

// hw_ prefix for "helloworld" because there might be overlap with tables from
// other projects in shared db

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"unique"`

	IDPUserID string `gorm:"uniqueIndex:hw_compositeidentity"`
	IDPIssuer string `gorm:"uniqueIndex:hw_compositeidentity"`

	Sessions []Session `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
}

func (User) TableName() string { return "hw_users" }

type Session struct {
	gorm.Model

	UserID      uint
	AccessToken string
	IDToken     string
}

func (Session) TableName() string { return "hw_sessions" }
