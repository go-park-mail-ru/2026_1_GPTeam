package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/grpcserver"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/httpserver"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/storage"
	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	debug := os.Getenv("DEBUG") == "true"
	if err := logger.InitLogger(debug); err != nil {
		fmt.Println("Error initializing logger:", err)
		return
	}
	defer func() { _ = logger.Close() }()

	log := logger.GetLogger()

	storageRoot := os.Getenv("FILESERVER_STORAGE_PATH")
	if storageRoot == "" {
		storageRoot = "./static"
	}
	if err := os.MkdirAll(storageRoot, 0o755); err != nil {
		log.Fatal("fileserver storage", zap.Error(err))
	}

	httpAddr := os.Getenv("FILESERVER_HTTP_LISTEN")
	if httpAddr == "" {
		httpAddr = ":8082"
	}
	grpcAddr := os.Getenv("FILESERVER_GRPC_LISTEN")
	if grpcAddr == "" {
		grpcAddr = ":50053"
	}

	metricsPort := os.Getenv("FILESERVER_METRICS_PORT")
	if metricsPort == "" {
		log.Fatal("FILESERVER_METRICS_PORT environment variable not set")
		return
	}
	metricsPort = ":" + metricsPort
	registry := prometheus.NewRegistry()
	metrics.InitMetrics(registry)
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	metricsServer := &http.Server{
		Addr:         metricsPort,
		Handler:      metricsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		log.Info("starting metrics", zap.String("addr", metricsPort))
		err := metricsServer.ListenAndServe()
		if err != nil {
			log.Fatal("Error starting metrics server", zap.Error(err))
			return
		}
	}()

	avatarStorage := storage.NewLocalStorage(storageRoot)
	avatarApp := application.NewAvatarService(avatarStorage)
	server := grpcserver.NewServer(avatarApp)

	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("fileserver grpc listen", zap.String("addr", grpcAddr), zap.Error(err))
	}
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(prometheusUnaryInterceptor))
	fsv1.RegisterFileServiceServer(grpcServer, server)

	go func() {
		log.Info("fileserver gRPC listening", zap.String("addr", grpcAddr))
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatal("fileserver grpc serve", zap.Error(err))
		}
	}()

	httpHandler := httpserver.NewRouter(storageRoot)
	httpSrv := &http.Server{
		Addr:         httpAddr,
		Handler:      httpHandler,
		ReadTimeout:  parseDurSec(os.Getenv("FILESERVER_READ_TIMEOUT_SEC"), 30*time.Second),
		WriteTimeout: parseDurSec(os.Getenv("FILESERVER_WRITE_TIMEOUT_SEC"), 120*time.Second),
	}

	log.Info("fileserver HTTP listening", zap.String("addr", httpAddr), zap.String("storage", storageRoot))
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal("fileserver http", zap.Error(err))
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

func prometheusUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	method := info.FullMethod
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	appMetrics := metrics.GetMetrics()
	appMetrics.FsGrpcRequestsDuration.WithLabelValues(method).Observe(float64(duration.Milliseconds()))
	statusCode := "OK"
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code().String()
		} else {
			statusCode = "Unknown"
		}
	}
	appMetrics.FsGrpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
	return resp, err
}
