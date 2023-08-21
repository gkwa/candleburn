package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init() {
	Logger, _ = GetLogger()
}

func GetLogger() (*zap.Logger, error) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeTime = nil // Removing timestamp from logs
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller())

	return logger, nil
}

func createLogger() (*zap.Logger, error) {
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"brownfish.log"}, // Update the file path here
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}

	return config.Build()
}

func Sync() {
	Logger.Sync()
}
