package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once
var logger *zap.Logger
var mu sync.RWMutex
var file *os.File
var initErr error

func InitLogger() error {
	once.Do(func() {
		file, initErr = os.OpenFile("backend.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if initErr != nil {
			initErr = fmt.Errorf("error opening log file: %w", initErr)
			return
		}
		fileWriter := zapcore.AddSync(file)
		consoleWriter := zapcore.AddSync(os.Stdout)
		logEncoderConfig := zap.NewProductionEncoderConfig()
		logEncoderConfig.TimeKey = "time"
		logEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logCore := zapcore.NewTee(
			zapcore.NewCore(zapcore.NewJSONEncoder(logEncoderConfig), fileWriter, zap.DebugLevel),
			zapcore.NewCore(zapcore.NewConsoleEncoder(logEncoderConfig), consoleWriter, zap.InfoLevel),
		)
		logger = zap.New(logCore, zap.AddCaller())
	})
	return initErr
}

func GetLogger() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	var multiErr error
	if logger != nil {
		errLogger := logger.Sync()
		logger = nil
		errFile := file.Close()
		file = nil
		multiErr = errors.Join(errLogger, errFile)
	}
	return multiErr
}

func GetLoggerWIthRequestId(ctx context.Context) *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	requestId, ok := ctx.Value("request_id").(string)
	if !ok {
		return logger
	}
	return logger.With(zap.String("request_id", requestId))
}
