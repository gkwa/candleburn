package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

func Must(logger *Logger, err error) *Logger {
	if err != nil {
		panic(err)
	}
	return logger
}

func NewLogger(logFile string) (*Logger, error) {
	config := zap.NewProductionConfig()
	
	defaultLogLevel := zapcore.DebugLevel
	
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeTime = nil

	consoleEncoder := zapcore.NewConsoleEncoder(config.EncoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}
	return &Logger{zap: logger}, err
}

func (l Logger) Debug(msg string, fields ...zap.Field) {
	l.writer().Debug(msg, fields...)
}

func (l Logger) Info(msg string, fields ...zap.Field) {
	l.writer().Info(msg, fields...)
}

func (l Logger) Error(msg string, fields ...zap.Field) {
	l.writer().Error(msg, fields...)
}

func (l Logger) Fatal(msg string, fields ...zap.Field) {
	l.writer().Error(msg, fields...)
}

var noOpLogger = zap.NewNop()

func (l Logger) writer() *zap.Logger {
	if l.zap == nil {
		return noOpLogger
	}
	return l.zap
}
