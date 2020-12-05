package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan 01, 2006  15:04:05"))
}

func CustomEncodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logLevelSeverity[level])
}

func CustomLevelFileEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + logLevelSeverity[level] + "]")
}

func newZap() *zap.Logger {
	var w zapcore.WriteSyncer

	if File != "" {
		w = zapcore.AddSync(&lumberjack.Logger{
			Filename:   File,
			MaxSize:    100,
			MaxBackups: 20,
			MaxAge:     28,
			Compress:   true,
		})
	}

	cfgConsole := zapcore.EncoderConfig{
		MessageKey:   "msg",
		LevelKey:     "level",
		TimeKey:      "time",
		EncodeLevel:  CustomEncodeLevel,
		EncodeTime:   SyslogTimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(d.Nanoseconds() / 1e6)
		},
		EncodeName: zapcore.FullNameEncoder,
	}

	cfgFile := zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		TimeKey:     "time",
		EncodeLevel: CustomLevelFileEncoder,
		EncodeTime:  SyslogTimeEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(d.Nanoseconds() / 1e6)
		},
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeName:   zapcore.FullNameEncoder,
	}
	consoleDebugging := zapcore.Lock(os.Stdout)

	var cores []zapcore.Core
	if File != "" {
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(cfgFile), w, zap.DebugLevel))
	}

	cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(cfgConsole), consoleDebugging, zap.DebugLevel))
	return zap.New(zapcore.NewTee(cores...)) //.WithOptions(zap.AddCaller(), zap.AddCallerSkip(2))
}
