package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	DEBUG := os.Getenv("DEBUG") == "true"
	if err := logger.InitLogger(DEBUG); err != nil {
		fmt.Println("Error initializing logger:", err)
		return
	}
	defer func() { _ = logger.Close() }()

	log := logger.GetLogger()

	storage := os.Getenv("FILESERVER_STORAGE_PATH")
	if storage == "" {
		storage = "./static"
	}
	if err := os.MkdirAll(storage, 0755); err != nil {
		log.Fatal("fileserver storage", zap.Error(err))
	}
	token := os.Getenv("FILESERVER_UPLOAD_TOKEN")

	addr := os.Getenv("FILESERVER_LISTEN")
	if addr == "" {
		addr = ":8082"
	}

	handler := fileserver.NewRouter(storage, token)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  parseDurSec(os.Getenv("FILESERVER_READ_TIMEOUT_SEC"), 30*time.Second),
		WriteTimeout: parseDurSec(os.Getenv("FILESERVER_WRITE_TIMEOUT_SEC"), 120*time.Second),
	}

	log.Info("fileserver listening", zap.String("addr", addr), zap.String("storage", storage))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("fileserver", zap.Error(err))
	}
}

func parseDurSec(s string, def time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	sec, err := strconv.Atoi(s)
	if err != nil || sec <= 0 {
		return def
	}
	return time.Duration(sec) * time.Second
}
