package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// My context key is of type string.
type myKey string

// Key's used for context values
const (
	TrackingIDKey = myKey("trackingid") // track id key.
	TenantIDKey   = myKey("tenantid")   // tenant id key.
	LocationIDKey = myKey("locationid") // location id key.
	UserIDKey     = myKey("userid")     // userid key.
	LoggerKey     = myKey("Logger")     // logger key
	ServiceName   = "ServiceName"
	LogType       = "LogType"
	PID           = "PID"

	// Constant to use as HTTP Header key for REST APIs
	// Webex and ControlHub use "TrackingID"  and thus making it consistent with it
	TrackingIDHeader = "TrackingID"
)

// DefaultConfig - default Development logger configuration
var DefaultConfig = &Config{
	Enabled:           true,
	Level:             "INFO",
	Development:       true,
	DisableCaller:     false,
	DisableStacktrace: false,
	Encoding:          "console",
	TimeKey:           "",
	MessageKey:        "msg",
	LevelKey:          "level",
	NameKey:           "logger",
	CallerKey:         "caller",
	StacktraceKey:     "stacktrace",
	CustomStacktrace:  false,
}

// DefaultProdConfig - default Production logger configuration
var DefaultProdConfig = &Config{
	Enabled:           true,
	Level:             "INFO",
	Development:       false,
	DisableCaller:     false,
	DisableStacktrace: false,
	Encoding:          "json",
	TimeKey:           "ts",
	MessageKey:        "msg",
	LevelKey:          "level",
	NameKey:           "logger",
	CallerKey:         "caller",
	StacktraceKey:     "stacktrace",
	CustomStacktrace:  false,
}

// DefaultMutedTestConfig - default unit test logger configuration for muting all logs
var DefaultMutedTestConfig = &Config{
	Enabled: false,
}

// Config struct contains configuration information for logging
type Config struct {
	Enabled           bool   `config:"env:LOG_ENABLED,yaml"`
	Level             string `config:"env:LOG_LEVEL,yaml"`
	Development       bool   `config:"env:LOG_DEVELOPMENT,yaml"`
	DisableCaller     bool   `config:"env:LOG_DISABLE_CALLER,yaml"`
	DisableStacktrace bool   `config:"env:LOG_DISABLE_STACK_TRACE,yaml"`
	Encoding          string `config:"env:LOG_ENCODING,yaml"`
	TimeKey           string `config:"env:LOG_TIME_KEY,yaml"`
	MessageKey        string `config:"env:LOG_MESSAGE_KEY,yaml"`
	LevelKey          string `config:"env:LOG_LEVEL_KEY,yaml"`
	NameKey           string `config:"env:LOG_NAME_KEY,yaml"`
	CallerKey         string `config:"env:LOG_CALLER_KEY,yaml"`
	StacktraceKey     string `config:"env:LOG_STACKTRACE_KEY,yaml"`
	CustomStacktrace  bool   `config:"env:LOG_CUSTOM_STACK_TRACE,yaml"`
}

// Logger embeds zap.Logger
// TODO: Add log methods taking context or add zap fields wrapping context.
type Logger struct {
	root             *zap.Logger
	log              *zap.Logger
	Config           zap.Config
	serviceName      string
	trackingIDPrefix string
	logConfig        *Config
}

// NewNop constructs a nop zap logger
func NewNop() *Logger {
	return &Logger{
		log:       zap.NewNop(),
		logConfig: DefaultConfig,
	}
}

// NewTrackingID generates a UUID Version 1 based trackingID for logging purpose.
// This is used to track the control flow through various services.
// Basing on UUID Version 1 allows map to timestamp and device mac address and track the logs.
func NewTrackingID(sender string) string {
	uuid1, err := uuid.NewUUID()
	if err != nil {
		// Version 1 failed - use Version 4 - timesatamp/macaddress unavailable
		uuid1 = uuid.New()
	}
	return fmt.Sprintf("%s_%s", sender, uuid1.String())
}

// NewTimeRandomTrackingID returns a new trackingID for the service
// UUID used is based on Time + Random
func NewTimeRandomTrackingID(serviceName string) string {
	return fmt.Sprintf("%s_%s", serviceName, NewUUID().String())
}

// NewTestLogger initializes global logger object to use Tracer in Tests
func NewTestLogger(serviceName string) (*Logger, func()) {
	logger, zapConfig := defaultLogger(serviceName)
	etiLogger := &Logger{
		log:              logger,
		Config:           zapConfig,
		serviceName:      serviceName,
		trackingIDPrefix: serviceName,
		logConfig:        DefaultConfig,
	}
	return etiLogger, stopper(logger)
}

// AddTrackingNameValuePairs adds Key:Value pairs to the TrackingID as per Cisco standard
// https://github.com/CiscoDevNet/api-design-guide#352-trackingid-header
func AddTrackingNameValuePairs(trackingID string, pairs ...string) (string, error) {
	nvs, err := mapFromPairsToString(pairs...)
	if err != nil {
		return trackingID, err
	}
	return fmt.Sprintf("%s%s", trackingID, nvs), nil
}

// checkPairs returns the count of strings passed in, and an error if the count is not an even number.
func checkPairs(pairs ...string) (int, error) {
	length := len(pairs)
	if length%2 != 0 {
		return length, fmt.Errorf("number of parameters must be multiple of 2, got %v", pairs)
	}
	return length, nil
}

// mapFromPairsToString converts variadic string parameters to a string.
func mapFromPairsToString(pairs ...string) (string, error) {
	length, err := checkPairs(pairs...)
	if err != nil {
		return "", err
	}
	nvs := ""
	for i := 0; i < length; i += 2 {
		nvs = fmt.Sprintf("%s_%s:%s", nvs, pairs[i], pairs[i+1])
	}
	return nvs, nil
}

// TrackingNameValuePairs returns Key:Value pairs from TrackingID as
// string to string map
func TrackingNameValuePairs(trackingID string) (map[string]string, error) {
	empty := make(map[string]string)
	parts := strings.Split(trackingID, "_")
	if len(parts) < 3 {
		return empty, fmt.Errorf("name:value pairs are not found in trackingID %s", trackingID)
	}

	pairs := make(map[string]string)
	for _, part := range parts[2:] {
		pair := strings.Split(part, ":")
		if len(pair) != 2 {
			return empty, fmt.Errorf("name:value wrong format %s", part)
		}
		pairs[pair[0]] = pair[1]
	}
	return pairs, nil
}

// TrackingIDTime returns the time encoded in TrackingID equal to
// the number of seconds and nanoseconds using the Unix epoch of 1 Jan 1970
func TrackingIDTime(trackingID string) (sec int64, nsec int64, err error) {
	parts := strings.Split(trackingID, "_")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("trackingID wrong format")
	}
	uuid1, err := uuid.Parse(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing TrackingID: %s", err.Error())
	}
	sec, nsec = uuid1.Time().UnixTime()
	return sec, nsec, nil
}

// TimedRandomTrackingIDTime returns the time encoded in TrackingID
func TimedRandomTrackingIDTime(trackingID string) (time.Time, error) {
	parts := strings.Split(trackingID, "_")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("trackingID wrong format")
	}

	u, err := FromString(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	t, err := u.GetTimeStamp()
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// Default create Default logger and set Common fields
func Default(serviceName string) (*Logger, func()) {
	root, zapConfig := defaultLogger(serviceName)
	// Set common log fields
	logger := root.With(
		zap.String(ServiceName, serviceName),
		zap.Int(LogType, int(zapConfig.Level.Level())),
		zap.Int(PID, os.Getpid()),
	)
	// Set ZAP global logger to log with common fields and our config
	zap.ReplaceGlobals(logger)

	etiLogger := &Logger{
		root:             root,
		log:              logger,
		Config:           zapConfig,
		serviceName:      serviceName,
		trackingIDPrefix: serviceName,
		logConfig:        DefaultConfig,
	}
	return etiLogger, stopper(logger)
}

// New create New logger and set Common fields
func New(serviceName string, config *Config) (*Logger, func(), error) {
	// Provide Noop logger if logging is disabled
	if !config.Enabled {
		return NewNop(), func() { /*NOP*/ }, nil
	}
	root, zapConfig, err := newLogger(serviceName, config)
	if err != nil {
		return nil, nil, err
	}
	// Set common log fields
	logger := root.With(
		zap.String(ServiceName, serviceName),
		zap.Int(LogType, int(zapConfig.Level.Level())),
		zap.Int(PID, os.Getpid()),
	)
	// Set ZAP global logger to log with common fields and our config
	zap.ReplaceGlobals(logger)

	etiLogger := &Logger{
		root:             root,
		log:              logger,
		Config:           zapConfig,
		serviceName:      serviceName,
		trackingIDPrefix: serviceName,
		logConfig:        config,
	}
	return etiLogger, stopper(logger), nil
}

func stopper(logger *zap.Logger) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic while syncing zap logger")
			}
		}()
		if logger != nil {
			_ = logger.Sync()
		}
	}
}

func defaultLogger(serviceName string) (*zap.Logger, zap.Config) {
	// Zap configuration
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.Encoding = DefaultConfig.Encoding
	zapConfig.DisableCaller = DefaultConfig.DisableCaller
	zapConfig.DisableStacktrace = true
	zapConfig.Sampling = nil
	// Encoding configuration
	zapConfig.EncoderConfig.TimeKey = "ts"
	zapConfig.EncoderConfig.MessageKey = DefaultConfig.MessageKey
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := zapConfig.Build()

	// skip 1 level of caller,  so logs aren't logging this interface calls
	logger = logger.WithOptions(zap.AddCallerSkip(1))

	// Set Log level to Info
	_ = zapConfig.Level.UnmarshalText([]byte(strings.ToLower("INFO")))

	return logger, zapConfig
}

func newLogger(serviceName string, config *Config) (*zap.Logger, zap.Config, error) {
	// Merge default with supplied configuration
	if err := mergo.Merge(config, DefaultProdConfig); err != nil {
		return nil, zap.Config{}, err
	}
	// Zap configuration
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.Encoding = config.Encoding
	zapConfig.DisableCaller = config.DisableCaller
	zapConfig.DisableStacktrace = config.DisableStacktrace
	zapConfig.Sampling = nil
	// Encoding configuration
	zapConfig.EncoderConfig.TimeKey = config.TimeKey
	zapConfig.EncoderConfig.MessageKey = config.MessageKey
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, zap.Config{}, err
	}

	// skip 1 level of caller,  so logs aren't logging this interface calls
	logger = logger.WithOptions(zap.AddCallerSkip(1))

	// Set Log level from configuration
	if config.Level != "" {
		if err := zapConfig.Level.UnmarshalText([]byte(strings.ToLower(config.Level))); err != nil {
			return nil, zap.Config{}, errors.Errorf("Unable to parse log level from config %s: %s", config.Level, err.Error())
		}
	}
	return logger, zapConfig, nil
}

// Zap return the underlying zap log writer
func (l *Logger) Zap() *zap.Logger {
	return l.log
}

// Close flushes the underlying logger causing it to submit any queued logs
func (l *Logger) Close() {
	_ = l.log.Sync()
}

// SetLevel sets log level. Can be: debug, info, warn, error, fatal
func (l *Logger) SetLevel(level string) error {
	err := l.Config.Level.UnmarshalText([]byte(strings.ToLower(level)))
	if err == nil {
		l.log = l.root.With(l.getCommonFields()...)
	}
	return err
}

// ====================================================================
//
// # Helper functions
//
// ====================================================================
// Function to populate the common fields alond with provided entry details.
func (l *Logger) populateCommonFields() *zap.Logger {
	return l.log.With(l.getCommonFields()...)
}

// populateContext takes context and log.entry and Returns Entry
func (l *Logger) populateContextFields(ctx context.Context) *zap.Logger {
	return l.log.With(getContextFields(ctx)...)
}

// populateContext takes context and log.entry and Returns Entry
func (l *Logger) populateContextCommonFields(ctx context.Context) *zap.Logger {
	f := append(getContextFields(ctx), l.getCommonFields()...)
	return l.log.With(f...)
}

// getCommonFields Returns Common zap.Fields
func (l *Logger) getCommonFields() []zapcore.Field {
	return []zapcore.Field{
		zap.String(ServiceName, l.serviceName),
		zap.Int(LogType, int(l.Config.Level.Level())),
		zap.Int(PID, os.Getpid()),
	}
}

// getContextFields takes context and Returns zap.Fields
func getContextFields(ctx context.Context) []zapcore.Field {
	keysAndValues := make([]zapcore.Field, 0)
	if ctx != nil {
		if trackingID := ctx.Value(TrackingIDKey); trackingID != nil {
			keysAndValues = append(keysAndValues, zap.String(string(TrackingIDKey), trackingID.(string)))
		}
		if tenantID := ctx.Value(TenantIDKey); tenantID != nil {
			keysAndValues = append(keysAndValues, zap.String(string(TenantIDKey), tenantID.(string)))
		}
		if locationID := ctx.Value(LocationIDKey); locationID != nil {
			keysAndValues = append(keysAndValues, zap.String(string(LocationIDKey), locationID.(string)))
		}
		if userID := ctx.Value(UserIDKey); userID != nil {
			keysAndValues = append(keysAndValues, zap.String(string(UserIDKey), userID.(string)))
		}
	}
	return keysAndValues
}

// NewTrackingID generates a UUID Version 1 based trackingID for logging purpose.
// This is used to track the control flow through various services.
// Basing on UUID Version 1 allows map to timestamp and device mac address and track the logs.
func (l *Logger) NewTrackingID() string {
	return NewTrackingID(l.trackingIDPrefix)
}

// NewTimeRandomTrackingID returns a new trackingID for the service
// UUID used is based on Time + Random
func (l *Logger) NewTimeRandomTrackingID() string {
	return NewTimeRandomTrackingID(l.trackingIDPrefix)
}

// SetTrackingIDPrefix sets prefix for tracking IDs generated
func (l *Logger) SetTrackingIDPrefix(prefix string) {
	l.trackingIDPrefix = prefix
}

// TrackingIDPrefix gets prefix of the generated tracking IDs
func (l *Logger) TrackingIDPrefix() string {
	return l.trackingIDPrefix
}

// WithContext returns logger with context fields set
// Usage:
// log.WithContext(ctx).Info(msg, i ...interface{})
// Example:
// log.WithContext(ctx).Info("redirect to %s", "http://foo/foo")
func (l *Logger) WithContext(ctx context.Context) *zap.SugaredLogger {
	return l.log.WithOptions(zap.AddCallerSkip(-1)).With(getContextFields(ctx)...).Sugar()
}

// WithContextDepth resets the depth of the caller and returns logger with context and additional key-value fields set
// Usage:
// log.WithContextFields(ctx, addDepth, fields ...interface{}).Info(msg, i ...interface{})
// Example:
// log.WithContextFields(ctx, 1, "code", 301).Info("redirect to %s", "http://foo/foo")
func (l *Logger) WithContextDepth(ctx context.Context, addDepth int, fields ...interface{}) *zap.SugaredLogger {
	var sugar *zap.SugaredLogger
	if addDepth == 1 {
		sugar = l.log.With(getContextFields(ctx)...).Sugar()
	} else {
		addDepth--
		sugar = l.log.WithOptions(zap.AddCallerSkip(addDepth)).With(getContextFields(ctx)...).Sugar()
	}
	if len(fields) > 0 {
		return sugar.With(fields...)
	}
	return sugar
}

// WithContextFields returns logger with context and additional key-value fields set
// Usage:
// log.WithContextFields(ctx, fields ...interface{}).Info(msg, i ...interface{})
// Example:
// log.WithContextFields(ctx, "code", 301).Info("redirect to %s", "http://foo/foo")
func (l *Logger) WithContextFields(ctx context.Context, fields ...interface{}) *zap.SugaredLogger {
	return l.log.WithOptions(zap.AddCallerSkip(-1)).With(getContextFields(ctx)...).Sugar().With(fields...)
}

// WithMap adds all map key-value pairs to the logging context.
func (l *Logger) WithMap(nameValue map[string]interface{}) *zap.SugaredLogger {
	if nameValue != nil {
		fields := make([]interface{}, 0)
		for k, v := range nameValue {
			if k == "" {
				k = "EMPTY"
			}
			fields = append(fields, checkEmpty(k, "EMPTY"), checkEmpty(v.(string), "EMPTY"))
		}
		return l.log.Sugar().With(fields...)
	}
	return l.log.Sugar()
}

func checkEmpty(s string, empty string) string {
	if s == "" {
		s = empty
	}
	return s
}

// WithFieldsz adds a variadic number of strongly typed Fields to the logging context.
func (l *Logger) WithFieldsz(fields ...zapcore.Field) *zap.Logger {
	return l.log.With(fields...)
}

// WithContextz adds TrackingIDKey, TenantIDKey, LocationIDKey and UserIDKey from the provided context as well as
// variadic number of strongly typed Fields to the logging context.
func (l *Logger) WithContextz(ctx context.Context, fields ...zapcore.Field) *zap.Logger {
	return l.log.With(getContextFields(ctx)...).With(fields...)
}

// SetFields adds provided variadic number of strongly typed Fields to the logging context of
// the logger object. Every log by that object will contain those fields moving forward.
func (l *Logger) SetFields(fields ...zapcore.Field) {
	l.log = l.log.With(fields...)
}

// SetContext adds TrackingIDKey, TenantIDKey, LocationIDKey and UserIDKey from the provided context to the
// logging context of the logger object.
// Every log by that object will contain those fields moving forward.
func (l *Logger) SetContext(ctx context.Context) {
	l.log = l.log.With(getContextFields(ctx)...)
}

// CloneWithContext clones logger, setting context fields
func (l *Logger) CloneWithContext(ctx context.Context) *Logger {
	return &Logger{
		root:        l.root,
		log:         l.populateContextFields(ctx),
		Config:      l.Config,
		serviceName: l.serviceName,
		logConfig:   l.logConfig,
	}
}

// New creates child logger, with fields
func (l *Logger) New(fields ...zapcore.Field) *Logger {
	return &Logger{
		root:        l.root,
		log:         l.log.With(fields...),
		Config:      l.Config,
		serviceName: l.serviceName,
		logConfig:   l.logConfig,
	}
}

func getStackTrace(key string) zapcore.Field {
	buf := make([]byte, 1<<16)
	sz := runtime.Stack(buf, false)
	if sz != 0 {
		return zap.String(key, formatStackTrace(buf[0:sz]))
	}
	return zap.String(key, "")
}

func formatStackTrace(stack []byte) string {
	//fmt.Printf("\n%s", stack)
	var out strings.Builder
	var buf []byte
	var b int
	a := bytes.Split(stack, []byte("\n"))
	for i := 5; i < len(a)-1; i += 2 {
		out.WriteString("[")
		buf = a[i+1]
		b = bytes.LastIndexByte(buf, byte('/'))
		if b > -1 {
			buf = buf[b+1:]
		}
		b = bytes.LastIndexByte(buf, byte(' '))
		if b > -1 {
			out.Write(buf[:b+1])
		}

		buf = a[i]
		b = bytes.LastIndexByte(buf, byte('.'))
		if b > -1 {
			out.Write(buf[b+1:])
		}
		out.WriteString("] ")
	}
	return out.String()
}

// ====================================================================
//
// Log functions
//
// ====================================================================

// Debug takes a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func (l *Logger) Debug(msg string, i ...interface{}) {
	l.log.Debug(fmt.Sprintf(msg, i...))
}

// Info takes a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func (l *Logger) Info(msg string, i ...interface{}) {
	l.log.Info(fmt.Sprintf(msg, i...))
}

// Error logs a message at error level,  setting a zapcore error field with the err
func (l *Logger) Error(msg string, i ...interface{}) {
	if l.logConfig != nil && l.logConfig.CustomStacktrace {
		l.log.With(getStackTrace(l.logConfig.StacktraceKey)).Error(fmt.Sprintf(msg, i...))
	} else {
		l.log.Error(fmt.Sprintf(msg, i...))
	}
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, i ...interface{}) {
	l.log.Warn(fmt.Sprintf(msg, i...))
}

// Fatal logs a fatal message and error, then exists
func (l *Logger) Fatal(msg string, i ...interface{}) {
	l.log.Fatal(fmt.Sprintf(msg, i...))
}

// Debugx takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func (l *Logger) Debugx(ctx context.Context, msg string, i ...interface{}) {
	l.log.With(getContextFields(ctx)...).Debug(fmt.Sprintf(msg, i...))
}

// Infox takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func (l *Logger) Infox(ctx context.Context, msg string, i ...interface{}) {
	l.log.With(getContextFields(ctx)...).Info(fmt.Sprintf(msg, i...))
}

// Errorx logs context and a message at error level,  setting a zapcore error field with the err
func (l *Logger) Errorx(ctx context.Context, msg string, i ...interface{}) {
	if l.logConfig != nil && l.logConfig.CustomStacktrace {
		l.log.With(getStackTrace(l.logConfig.StacktraceKey)).With(getContextFields(ctx)...).Error(fmt.Sprintf(msg, i...))
	} else {
		l.log.With(getContextFields(ctx)...).Error(fmt.Sprintf(msg, i...))
	}
}

// Warnx logs context and a message at warn level
func (l *Logger) Warnx(ctx context.Context, msg string, i ...interface{}) {
	l.log.With(getContextFields(ctx)...).Warn(fmt.Sprintf(msg, i...))
}

// Fatalx logs context and a fatal message and error, then exists
func (l *Logger) Fatalx(ctx context.Context, msg string, i ...interface{}) {
	l.log.With(getContextFields(ctx)...).Fatal(fmt.Sprintf(msg, i...))
}

// WithContext returns logger with context fields set
// Usage:
// log.WithContext(ctx).Info(msg, i ...interface{})
// Example:
// log.WithContext(ctx).Info("redirect to %s", "http://foo/foo")
func WithContext(ctx context.Context) *zap.SugaredLogger {
	return zap.L().WithOptions(zap.AddCallerSkip(-1)).With(getContextFields(ctx)...).Sugar()
}

// WithContextDepth resets the depth of the caller and returns logger with context and additional key-value fields set
// Usage:
// log.WithContextFields(ctx, addDepth, fields ...interface{}).Info(msg, i ...interface{})
// Example:
// log.WithContextFields(ctx, 1, "code", 301).Info("redirect to %s", "http://foo/foo")
func WithContextDepth(ctx context.Context, addDepth int, fields ...interface{}) *zap.SugaredLogger {
	var sugar *zap.SugaredLogger
	if addDepth == 1 {
		sugar = zap.L().With(getContextFields(ctx)...).Sugar()
	} else {
		addDepth--
		sugar = zap.L().WithOptions(zap.AddCallerSkip(addDepth)).With(getContextFields(ctx)...).Sugar()
	}
	if len(fields) > 0 {
		return sugar.With(fields...)
	}
	return sugar
}

// WithContextFields returns logger with context and additional key-value fields set
// Usage:
// log.WithContextFields(ctx, fields ...interface{}).Info(msg, i ...interface{})
// Example:
// log.WithContextFields(ctx, "code", 301).Info("redirect to %s", "http://foo/foo")
func WithContextFields(ctx context.Context, fields ...interface{}) *zap.SugaredLogger {
	return zap.L().WithOptions(zap.AddCallerSkip(-1)).With(getContextFields(ctx)...).Sugar().With(fields...)
}

// DebugX takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func DebugX(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Debug(fmt.Sprintf(msg, i...))
}

// InfoX takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func InfoX(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Info(fmt.Sprintf(msg, i...))
}

// ErrorX logs context and a message at error level,  setting a zapcore error field with the err
func ErrorX(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Error(fmt.Sprintf(msg, i...))
}

// WarnX logs context and a message at warn level
func WarnX(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Warn(fmt.Sprintf(msg, i...))
}

// FatalX logs context and a fatal message and error, then exists
func FatalX(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Fatal(fmt.Sprintf(msg, i...))
}

// Debugx takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func Debugx(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Debug(fmt.Sprintf(msg, i...))
}

// Infox takes context and a msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func Infox(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Info(fmt.Sprintf(msg, i...))
}

// Errorx logs context and a message at error level,  setting a zapcore error field with the err
func Errorx(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Error(fmt.Sprintf(msg, i...))
}

// Warnx logs context and a message at warn level
func Warnx(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Warn(fmt.Sprintf(msg, i...))
}

// Fatalx logs context and a fatal message and error, then exists
func Fatalx(ctx context.Context, msg string, i ...interface{}) {
	zap.L().With(getContextFields(ctx)...).Fatal(fmt.Sprintf(msg, i...))
}

// Debug takes msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func Debug(msg string, i ...interface{}) {
	zap.L().Debug(fmt.Sprintf(msg, i...))
}

// Info takes msg (can be formatted ex: "%s was trying to be used,  err %s") and logs it at info level
func Info(msg string, i ...interface{}) {
	zap.L().Info(fmt.Sprintf(msg, i...))
}

// Error logs message at error level,  setting a zapcore error field with the err
func Error(msg string, i ...interface{}) {
	zap.L().Error(fmt.Sprintf(msg, i...))
}

// Warn logs message at warn level
func Warn(msg string, i ...interface{}) {
	zap.L().Warn(fmt.Sprintf(msg, i...))
}

// Fatal logs fatal message and error, then exists
func Fatal(msg string, i ...interface{}) {
	zap.L().Fatal(fmt.Sprintf(msg, i...))
}

// HTTPLogger is a middleware log helper
func HTTPLogger(logger *Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := httpsnoop.CaptureMetrics(next, w, r)
			logger.log.Info("access",
				zap.String("host", checkEmpty(checkEmpty(r.Header.Get("x-forwarded-for"), r.RemoteAddr), "-")),
				zap.String("ident", "-"),
				zap.String("time", time.Now().Format("2006-01-02 15:04:05.999Z")),
				zap.String("query_params", r.URL.RawQuery),
				zap.String("protocol", r.Proto),
				zap.String("user-name", "-"),
				zap.String("content-length", checkEmpty(r.Header.Get("Content-Length"), "-")),
				zap.String("user-agent", checkEmpty(r.Header.Get("user-agent"), "-")),
				zap.String("referer", checkEmpty(r.Header.Get("referer"), "-")),
				zap.Int("status", response.Code),
				zap.String("request-method", r.Method),
				zap.String("request-string", r.URL.Path),
				zap.Duration("response-time", response.Duration))
		})
	}
}

// LumberConfig config for Lumberjack logger
type LumberConfig struct {
	Filename   string `config:"def:out.log,env:LUMBER_FILENAME,yaml"`
	MaxSize    int    `config:"def:3,env:LUMBER_MAX_SIZE,yaml"` // megabytes
	MaxBackups int    `config:"def:1,env:LUMBER_MAX_BACKUPS,yaml"`
	MaxAge     int    `config:"def:7,env:LUMBER_MAX_AGE,yaml"` // days
	Compress   bool   `config:"def:false,false,env:LUMBER_COMPRESS,yaml"`
}

// DefaultLumberConfig - default Lumberjack configuration
var DefaultLumberConfig = &LumberConfig{
	Filename:   "out.log",
	MaxSize:    3,
	MaxBackups: 1,
	MaxAge:     7,
	Compress:   false,
}

// Syncer create New logger and set Common fields
// Use this instead of New() when you need to
// 1 - log to a file
// 2 - log to multiple destinations
//
// Usage:
//
// No console log - only log to file using Lumberjack config
//
//	:= Syncer("test", config, lumberConfig)
//
// Log to file and console
// := Syncer("test", config, lumberConfig, os.Stdout)
//
// Use plain text encoding (default is json)
//
//	config.Encoding = "console"
//	:= Syncer("test", config, lumberConfig, os.Stdout)
//
// Setup multiple destinations:
// - Stdout writer
// - Lumberjack writer
//
//	lumberWriter := &lumberjack.Logger{
//		Filename:   lumberConfig.Filename,
//		MaxSize:    lumberConfig.MaxSize,    // 3 (MB)                                                                                                                     // megabytes
//		MaxBackups: lumberConfig.MaxBackups, // 1
//		MaxAge:     lumberConfig.MaxAge,     // days
//		Compress:   lumberConfig.Compress,   // disabled by default
//	}
//
// Use multiple writers (no Lumberjack config - use lumberWriter directly)
//
//	:= Syncer("test", config, nil, lumberWriter, os.Stdout)
func Syncer(serviceName string, config *Config, lumberConfig *LumberConfig, writers ...io.Writer) (*Logger, func(), error) {
	// Provide Noop logger if logging is disabled
	if !config.Enabled {
		return NewNop(), func() { /*NOP*/ }, nil
	}
	root, zapConfig, err := syncerLogger(serviceName, config, lumberConfig, writers...)
	if err != nil {
		return nil, nil, err
	}
	// Set common log fields
	logger := root.With(
		zap.String(ServiceName, serviceName),
		zap.Int(LogType, int(zapConfig.Level.Level())),
		zap.Int(PID, os.Getpid()),
	)
	// Set ZAP global logger to log with common fields and our config
	zap.ReplaceGlobals(logger)

	etiLogger := &Logger{
		root:             root,
		log:              logger,
		Config:           zapConfig,
		serviceName:      serviceName,
		trackingIDPrefix: serviceName,
		logConfig:        config,
	}
	return etiLogger, stopper(logger), nil
}

func syncerLogger(serviceName string, config *Config, lumberConfig *LumberConfig, writers ...io.Writer) (*zap.Logger, zap.Config, error) {
	// Merge default with supplied configuration
	if err := mergo.Merge(config, DefaultProdConfig); err != nil {
		return nil, zap.Config{}, err
	}
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.Encoding = config.Encoding
	zapConfig.DisableCaller = config.DisableCaller
	zapConfig.DisableStacktrace = config.DisableStacktrace
	zapConfig.Sampling = nil
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = config.TimeKey
	encoderConfig.MessageKey = config.MessageKey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig = encoderConfig

	atomicLevel := zap.NewAtomicLevel()
	if err := atomicLevel.UnmarshalText([]byte(strings.ToLower(config.Level))); err != nil {
		return nil, zap.Config{}, err
	}

	zws := make([]zapcore.WriteSyncer, 0)
	if lumberConfig != nil {
		zws = append(zws, zapcore.AddSync(&lumberjack.Logger{
			Filename:   lumberConfig.Filename,
			MaxSize:    lumberConfig.MaxSize,    // 3 (MB)                                                                                                                     // megabytes
			MaxBackups: lumberConfig.MaxBackups, // 1
			MaxAge:     lumberConfig.MaxAge,     // days
			Compress:   lumberConfig.Compress,   // disabled by default
		}))
	}
	if len(writers) > 0 {
		for i := range writers {
			zws = append(zws, zapcore.AddSync(writers[i]))
		}
	}

	var encoder zapcore.Encoder
	if config.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zws...),
		atomicLevel,
	)
	logger := zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller(), zap.AddCallerSkip(1))
	return logger, zapConfig, nil
}
