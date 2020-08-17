package log

import (
	"go.uber.org/zap"
)

var (
	File      string
	DevMode   = true
	DebugMode = false
)

func Field(key string, value interface{}) field {
	return field{
		Key:   key,
		Value: value,
	}
}

func Err(err error) field {
	return field{
		Key:   "error",
		Value: err,
	}
}

type field struct {
	Key   string
	Value interface{}
}

var logger *zap.Logger

func getLogger() Logger {
	if logger == nil {
		logger = newZap()
		defer logger.Sync()

		return &defaultLogger{logger}
	}
	return &defaultLogger{logger}
}

// Logger is a convenience for logging
type Logger interface {
	Named(string) Logger
	Info(msg string, fields ...field)
	Debug(msg string, fields ...field)
	Warning(msg string, fields ...field)
	Error(msg string, fields ...field)
	Panic(msg string, fields ...field)
	Fatal(msg string, fields ...field)
}

type defaultLogger struct {
	zapper *zap.Logger
}

func (ul *defaultLogger) Named(name string) Logger {
	return &defaultLogger{zapper: ul.zapper.Named(name)}
}

func (ul *defaultLogger) Info(msg string, fields ...field) {
	var items []zap.Field
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Info(msg, items...)
}
func (ul *defaultLogger) Debug(msg string, fields ...field) {
	if DebugMode {
		var items []zap.Field
		for _, f := range fields {
			items = append(items, zap.Any(f.Key, f.Value))
		}
		ul.zapper.Debug(msg, items...)
	}
}
func (ul *defaultLogger) Warning(msg string, fields ...field) {
	if DebugMode {
		var items []zap.Field
		for _, f := range fields {
			items = append(items, zap.Any(f.Key, f.Value))
		}
		ul.zapper.Warn(msg, items...)
	}
}
func (ul *defaultLogger) Error(msg string, fields ...field) {
	var items []zap.Field
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Error(msg, items...)
}
func (ul *defaultLogger) Panic(msg string, fields ...field) {
	var items []zap.Field
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Panic(msg, items...)
}
func (ul *defaultLogger) Fatal(msg string, fields ...field) {
	var items []zap.Field
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Fatal(msg, items...)
}

func Named(name string) Logger {
	return getLogger().Named(name)
}

func Info(msg string, fields ...field) {
	getLogger().Info(msg, fields...)
}

// Debug used for showing more detailed activity
func Debug(msg string, fields ...field) {
	getLogger().Debug(msg, fields...)
}

func Warning(msg string, fields ...field) {
	getLogger().Warning(msg, fields...)
}

// Error displays more detailed error message
func Error(msg string, fields ...field) {
	getLogger().Error(msg, fields...)
}

func Panic(msg string, fields ...field) {
	getLogger().Panic(msg, fields...)
}

// Error displays more detailed error message
func Fatal(msg string, fields ...field) {
	getLogger().Fatal(msg, fields...)
}
