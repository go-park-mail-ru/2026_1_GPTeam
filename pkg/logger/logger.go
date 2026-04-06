package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

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
	if logger == nil {
		return zap.NewNop()
	}
	requestId, ok := ctx.Value("request_id").(string)
	if !ok {
		return logger
	}
	return logger.With(zap.String("request_id", requestId))
}

func ModifyLoggerWithDBQuery(log *zap.Logger, query string, args []any, duration time.Duration) *zap.Logger {
	return log.With(zap.String("query", query),
		zap.Any("args", args),
		zap.String("duration", duration.String()),
	)
}

var accessLogger *zap.Logger
var accessOnce sync.Once
var accessMu sync.RWMutex
var accessInitErr error
var accessFile *os.File

func InitAccessLogger() error {
	accessOnce.Do(func() {
		accessFile, accessInitErr = os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if accessInitErr != nil {
			accessInitErr = fmt.Errorf("error opening log file: %w", accessInitErr)
			return
		}
		fileWriter := zapcore.AddSync(accessFile)
		logEncoderConfig := zap.NewProductionEncoderConfig()
		logEncoderConfig.TimeKey = "time"
		logEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logCore := zapcore.NewCore(zapcore.NewJSONEncoder(logEncoderConfig), fileWriter, zap.DebugLevel)
		accessLogger = zap.New(logCore)
	})
	return accessInitErr
}

func GetAccessLogger() *zap.Logger {
	accessMu.RLock()
	defer accessMu.RUnlock()
	if accessLogger == nil {
		return zap.NewNop()
	}
	return accessLogger
}

func AccessClose() error {
	accessMu.Lock()
	defer accessMu.Unlock()
	var multiErr error
	if accessLogger != nil {
		errLogger := accessLogger.Sync()
		accessLogger = nil
		errFile := accessFile.Close()
		accessFile = nil
		multiErr = errors.Join(errLogger, errFile)
	}
	return multiErr
}
