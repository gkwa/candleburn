package logging

import "go.uber.org/zap"

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
	zap.NewProductionConfig()
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", "./logs/" + logFile}
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
	l.writer().Fatal(msg, fields...)
}

var noOpLogger = zap.NewNop()

func (l Logger) writer() *zap.Logger {
	if l.zap == nil {
		return noOpLogger
	}
	return l.zap
}
