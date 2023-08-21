package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeTime = nil // Removing timestamp from logs
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	Logger = zap.New(core, zap.AddCaller())
}
