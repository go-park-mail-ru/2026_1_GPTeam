package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/grpcserver"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	DEBUG := os.Getenv("DEBUG") == "true"
	err = logger.InitLogger(DEBUG)
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		return
	}
	defer func() {
		_ = logger.Close()
	}()

	log := logger.GetLogger()

	metricsPort := os.Getenv("AUTH_METRICS_PORT")
	if metricsPort == "" {
		log.Fatal("AUTH_METRICS_PORT environment variable not set")
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
		err = metricsServer.ListenAndServe()
		if err != nil {
			log.Fatal("Error starting metrics server", zap.Error(err))
			return
		}
	}()

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to create pool", zap.Error(err))
		return
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	jwtPostgres := repository.NewJwtPostgres(pool)
	jwtService, err := jwt_auth.NewJwt(jwtPostgres, os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
	if err != nil {
		log.Fatal("Failed to create JWT service", zap.Error(err))
		return
	}

	lisAddr := os.Getenv("AUTH_GRPC_LISTEN")
	if lisAddr == "" {
		lisAddr = ":50051"
	}
	lis, err := net.Listen("tcp", lisAddr)
	if err != nil {
		log.Fatal("auth grpc listen", zap.String("addr", lisAddr), zap.Error(err))
		return
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(prometheusUnaryInterceptor))
	authv1.RegisterAuthServiceServer(s, &grpcserver.Server{JWT: jwtService})

	log.Info("auth gRPC server listening", zap.String("addr", lisAddr))
	err = s.Serve(lis)
	if err != nil {
		log.Fatal("auth grpc serve", zap.Error(err))
	}
}

func prometheusUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	method := info.FullMethod
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	appMetrics := metrics.GetMetrics()
	appMetrics.AuthGrpcRequestsDuration.WithLabelValues(method).Observe(float64(duration.Milliseconds()))
	statusCode := "OK"
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code().String()
		} else {
			statusCode = "Unknown"
		}
	}
	appMetrics.AuthGrpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
	return resp, err
}
