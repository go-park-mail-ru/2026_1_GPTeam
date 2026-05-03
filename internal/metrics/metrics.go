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
	)
}
