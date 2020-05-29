package log

import (
	"go.uber.org/zap"
)

var (
	File string
	DevMode = true
	DebugMode = false
)

func Field(key string, value interface{}) field {
	return field{
		Key:   key,
		Value: value,
	}
}

type field struct {
	Key string
	Value interface{}
}

var logger *zap.Logger

var (
	//serv, erro, info, success, deb, txt func(txt string) string
)

func getLogger() Logger {
	if logger == nil {
		/*err := zap.RegisterSink("files", func(url *url.URL) (zap.Sink, error) {
			filename := futils.UnNormalizePath(url.Path)
			return os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
		})

		var logConfig zap.Config
		if DevMode {
			logConfig = zap.NewDevelopmentConfig()
		} else {
			logConfig = zap.NewProductionConfig()
			logConfig.OutputPaths = append(logConfig.OutputPaths, "stdout")
			if File != "" {
				logConfig.OutputPaths = append(logConfig.OutputPaths, fmt.Sprintf("files://%s", futils.NormalizePath(File)))
			}
		}

		logger, err = logConfig.Build(zap.AddCaller(), zap.AddCallerSkip(2), zap.WrapCore(func(core zapcore.Core) zapcore.Core {

			return core
		}))
		if err != nil {
			logger = zap.L()
		}

		zap.ReplaceGlobals(logger)*/
		logger = newZap()
		defer logger.Sync()


		return &defaultLogger{logger}
	}
	return &defaultLogger{logger}
}

// Logger is a convenience for logging
type Logger interface {
	in(msg string, fields ...field)
	de(msg string, fields ...field)
	er(msg string, err error, fields ...field)
	fa(msg string, err error, fields ...field)
}

type defaultLogger struct{
	zapper *zap.Logger
}

func (ul *defaultLogger) in(msg string, fields ...field) {
	var items []zap.Field
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Info(msg, items...)
}
func (ul *defaultLogger) de(msg string, fields ...field) {
	if DebugMode {
		var items []zap.Field
		for _, f := range fields {
			items = append(items, zap.Any(f.Key, f.Value))
		}
		ul.zapper.Debug(msg, items...)
	}
}
func (ul *defaultLogger) er(msg string, err error, fields ...field) {
	items := []zap.Field {zap.Error(err)}
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Error(msg, items...)
}
func (ul *defaultLogger) fa(msg string, err error, fields ...field) {
	items := []zap.Field {zap.Error(err)}
	for _, f := range fields {
		items = append(items, zap.Any(f.Key, f.Value))
	}
	ul.zapper.Fatal(msg, items...)
}


func Info(msg string, fields ...field) {
	getLogger().in(msg, fields...)
}

// Debug used for showing more detailed activity
func Debug(msg string, fields ...field) {
	getLogger().de(msg, fields...)
}

// Error displays more detailed error message
func Error(msg string, err error, fields ...field) {
	getLogger().er(msg, err, fields...)
}

// Error displays more detailed error message
func Fatal(msg string, err error, fields ...field) {
	getLogger().fa(msg, err, fields...)
}