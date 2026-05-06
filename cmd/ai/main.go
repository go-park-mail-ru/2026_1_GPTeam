package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/groq"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/grpcserver"
	aiv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	err := logger.InitLogger(false)
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		return
	}
	defer func() {
		err = logger.Close()
		if err != nil {
			fmt.Println("Error closing logger: ", err)
		}
	}()
	log := logger.GetLogger()

	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	if groqKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}
	proxyURLStr := os.Getenv("PROXY_URL")

	log.Info("groq api key loaded",
		zap.Int("len", len(groqKey)),
		zap.String("prefix", groqKey[:min(8, len(groqKey))]+"..."),
		zap.String("suffix", "..."+groqKey[max(0, len(groqKey)-4):]),
		zap.String("proxy", proxyURLStr),
	)

	metricsPort := os.Getenv("AI_METRICS_PORT")
	if metricsPort == "" {
		log.Fatal("AI_METRICS_PORT environment variable is required")
		return
	}
	metricsPort = ":" + metricsPort
	registry := prometheus.NewRegistry()
	metrics.InitMetrics(registry)
	registry.MustRegister(collectors.NewGoCollector())
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

	listenAddr := os.Getenv("AI_GRPC_LISTEN")
	if listenAddr == "" {
		listenAddr = ":50052"
	}

	groqClient := groq.NewGroqClient(groqKey, proxyURLStr)
	aiService := ai.NewGroqAiService(groqClient)

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(prometheusUnaryInterceptor))
	aiv1.RegisterAiServiceServer(grpcServer, &grpcserver.Server{AI: aiService})

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.String("addr", listenAddr), zap.Error(err))
	}

	log.Info("AI gRPC server starting", zap.String("addr", listenAddr))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down AI gRPC server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("AI gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Warn("AI gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func prometheusUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	method := info.FullMethod
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	appMetrics := metrics.GetMetrics()
	statusCode := "OK"
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code().String()
		} else {
			statusCode = "Unknown"
		}
	}
	appMetrics.AiGrpcRequestsDuration.WithLabelValues(method, statusCode).Observe(float64(duration.Milliseconds()))
	appMetrics.AiGrpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
	return resp, err
}
