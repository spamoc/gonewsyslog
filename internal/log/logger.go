package log

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logKey = "logger"

func levelEnableVerbose(l zapcore.Level) bool {
	return map[zapcore.Level]bool{
		zapcore.DebugLevel: true,
		zapcore.InfoLevel:  true,
		zapcore.WarnLevel:  true,
		zapcore.ErrorLevel: true,
		zapcore.FatalLevel: true,
		zapcore.PanicLevel: true,
	}[l]
}

func levelEnableLazy(l zapcore.Level) bool {
	return map[zapcore.Level]bool{
		zapcore.DebugLevel: false,
		zapcore.InfoLevel:  true,
		zapcore.WarnLevel:  true,
		zapcore.ErrorLevel: true,
		zapcore.FatalLevel: true,
		zapcore.PanicLevel: true,
	}[l]
}

func Get(ctx context.Context) *zap.Logger {
	a := ctx.Value(logKey)
	if a == nil {
		return Create(false)
	}
	return a.(*zap.Logger)
}

func Set(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)

}

func Bool(key string, p bool) zapcore.Field {
	return zap.Bool(key, p)
}

func Int(key string, p int) zapcore.Field {
	return zap.Int(key, p)
}

func Int16(key string, p int16) zapcore.Field {
	return zap.Int16(key, p)
}

func String(key string, p string) zapcore.Field {
	return zap.String(key, p)
}

func Uint(key string, p uint) zapcore.Field {
	return zap.Uint(key, p)
}

func Error(p error) zapcore.Field {
	return zap.Error(p)
}

func Any(key string, p interface{}) zapcore.Field {
	return zap.Any(key, p)
}

func Create(verbose bool) *zap.Logger {
	var levelEnabler zap.LevelEnablerFunc = nil
	var options []zap.Option = make([]zap.Option, 0)
	if verbose {
		levelEnabler = levelEnableVerbose
		// if verbose, write out stacktrace lazy log
		options = append(options, zap.AddStacktrace(zap.ErrorLevel))
	} else {
		levelEnabler = levelEnableLazy
	}
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(
				zapcore.EncoderConfig{
					MessageKey:     "msg",
					LevelKey:       "level",
					NameKey:        "name",
					TimeKey:        "ts",
					CallerKey:      "caller",
					StacktraceKey:  "stacktrace",
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeLevel:    zapcore.LowercaseLevelEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
				},
			),
			zapcore.AddSync(os.Stderr),
			levelEnabler,
		),
		options...,
	)
	return logger
}
