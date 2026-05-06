package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type AppMetrics struct {
	HttpRequestsTotal     *prometheus.CounterVec
	HttpRequestDuration   *prometheus.HistogramVec
	ActiveUsers           prometheus.Gauge
	DbQueryDuration       *prometheus.HistogramVec
	SupportCreationsTotal *prometheus.CounterVec
	// Auth microservice
	AuthGrpcRequestsTotal         *prometheus.CounterVec
	AuthGrpcRequestsDuration      *prometheus.HistogramVec
	AuthValidateTokenTotal        *prometheus.CounterVec
	AuthValidateRefreshTokenTotal *prometheus.CounterVec
	// FS microservice
	FsGrpcRequestsTotal    *prometheus.CounterVec
	FsGrpcRequestsDuration *prometheus.HistogramVec
	FsAvatarUploadDuration *prometheus.HistogramVec
	// AI microservice
	AiGrpcRequestsTotal              *prometheus.CounterVec
	AiGrpcRequestsDuration           *prometheus.HistogramVec
	AiGroqTranscribeRequestsDuration *prometheus.HistogramVec
	AiGroqParseRequestsDuration      *prometheus.HistogramVec
}

var once sync.Once
var mu sync.RWMutex
var metrics *AppMetrics

func InitMetrics(registry *prometheus.Registry) {
	once.Do(func() {
		metrics = &AppMetrics{
			HttpRequestsTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "app_http_requests_total",
					Help: "Общее количество HTTP-запросов",
				},
				[]string{"method", "endpoint", "status"},
			),
			HttpRequestDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "app_http_request_duration_milliseconds",
					Help:    "Длительность HTTP-запроса в миллисекундах",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"method", "endpoint"},
			),
			ActiveUsers: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "app_active_users",
					Help: "Текущее количество активных пользователей",
				},
			),
			DbQueryDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "app_db_query_duration_milliseconds",
					Help:    "Длительность запроса к базе данных в миллисекундах",
					Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
				},
				[]string{"query", "table"},
			),
			SupportCreationsTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "support_creations_total",
					Help: "Общее количество заявок в техподдержку",
				},
				[]string{},
			),
			// Auth microservice
			AuthGrpcRequestsTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "auth_grpc_requests_total",
					Help: "Общее количество gRPC запросов к микросервису авторизации",
				},
				[]string{"method", "status"},
			),
			AuthGrpcRequestsDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "auth_grpc_request_duration_milliseconds",
					Help:    "Длительность gRPC запросов к микросервису авторизации в миллисекундах",
					Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
				},
				[]string{"method"},
			),
			AuthValidateTokenTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "auth_validate_token_total",
					Help: "Общее количество проверок access token",
				},
				[]string{"valid"},
			),
			AuthValidateRefreshTokenTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "auth_validate_refresh_token_total",
					Help: "Общее количество проверок refresh token",
				},
				[]string{"valid"},
			),
			// FS microservice
			FsGrpcRequestsTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "fs_grpc_requests_total",
					Help: "Общее количество gRPC запросов к файловому микросервису",
				},
				[]string{"method", "status"},
			),
			FsGrpcRequestsDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "fs_grpc_request_duration_milliseconds",
					Help:    "Длительность gRPC запросов к файловому микросервису в миллисекундах",
					Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
				},
				[]string{"method"},
			),
			FsAvatarUploadDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: "fs_avatar_upload_duration_milliseconds",
					Help: "Общее время загрузки аватара в миллисекундах",
				},
				[]string{"status"},
			),
			// AI microservice
			AiGrpcRequestsTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "ai_grpc_requests_total",
					Help: "Общее количество gRPC запросов к AI микросервису",
				},
				[]string{"method", "status"},
			),
			AiGrpcRequestsDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "ai_grpc_request_duration_milliseconds",
					Help:    "Длительность gRPC запросов к AI микросервису в миллисекундах",
					Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
				},
				[]string{"method"},
			),
			AiGroqTranscribeRequestsDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: "ai_groq_transcribe_http_requests_duration_milliseconds",
					Help: "Длительность HTTP запросов на конвертацию голоса в текст к Groq API в миллисекундах",
				},
				[]string{"status"},
			),
			AiGroqParseRequestsDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: "ai_groq_parse_http_requests_duration_milliseconds",
					Help: "Длительность HTTP запросов на парсинг текста к Groq API в миллисекундах",
				},
				[]string{"status"},
			),
		}
		metrics.register(registry)
	})
}

func GetMetrics() *AppMetrics {
	mu.RLock()
	defer mu.RUnlock()
	if metrics == nil {
		return NewNoopMetrics()
	}
	return metrics
}

func NewNoopMetrics() *AppMetrics {
	return &AppMetrics{
		HttpRequestsTotal:     newNoopCounterVec(),
		HttpRequestDuration:   newNoopHistogramVec(),
		ActiveUsers:           newNoopGauge(),
		DbQueryDuration:       newNoopHistogramVec(),
		SupportCreationsTotal: newNoopCounterVec(),
	}
}

func newNoopGauge() prometheus.Gauge {
	return prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "noop"},
	)
}

func newNoopHistogramVec() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "noop"},
		[]string{},
	)
}

func newNoopCounterVec() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "noop"},
		[]string{},
	)
}

func (obj *AppMetrics) register(registry *prometheus.Registry) {
	registry.MustRegister(
		obj.HttpRequestsTotal,
		obj.HttpRequestDuration,
		obj.ActiveUsers,
		obj.DbQueryDuration,
		obj.SupportCreationsTotal,
		obj.AuthGrpcRequestsTotal,
		obj.AuthGrpcRequestsDuration,
		obj.AuthValidateTokenTotal,
		obj.AuthValidateRefreshTokenTotal,
		obj.FsGrpcRequestsTotal,
		obj.FsGrpcRequestsDuration,
		obj.FsAvatarUploadDuration,
		obj.AiGrpcRequestsTotal,
		obj.AiGrpcRequestsDuration,
		obj.AiGroqTranscribeRequestsDuration,
		obj.AiGroqParseRequestsDuration,
	)
}
