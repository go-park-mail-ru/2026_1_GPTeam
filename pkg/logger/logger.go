package logger

import (
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

func InitLogger() error {
	var err error
	once.Do(func() {
		file, err = os.OpenFile("backend.log", os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			err = fmt.Errorf("error opening log file: %w", err)
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
	return err
}

func GetLogger() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return logger
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	var err error
	if logger != nil {
		err = logger.Sync()
		logger = nil
		err = file.Close()
		file = nil
	}
	return err
}
