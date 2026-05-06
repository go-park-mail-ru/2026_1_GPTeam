package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewNoopMetrics(t *testing.T) {
	t.Parallel()

	m := NewNoopMetrics()
	require.NotNil(t, m)
	require.NotNil(t, m.HttpRequestsTotal)
	require.NotNil(t, m.HttpRequestDuration)
	require.NotNil(t, m.ActiveUsers)
	require.NotNil(t, m.DbQueryDuration)
	require.NotNil(t, m.SupportCreationsTotal)
	require.NotNil(t, m.AuthGrpcRequestsTotal)
	require.NotNil(t, m.AuthGrpcRequestsDuration)
	require.NotNil(t, m.AuthValidateTokenTotal)
	require.NotNil(t, m.AuthValidateRefreshTokenTotal)
	require.NotNil(t, m.FsGrpcRequestsTotal)
	require.NotNil(t, m.FsGrpcRequestsDuration)
	require.NotNil(t, m.FsAvatarUploadDuration)
	require.NotNil(t, m.AiGrpcRequestsTotal)
	require.NotNil(t, m.AiGrpcRequestsDuration)
	require.NotNil(t, m.AiGroqRequestsDuration)

	// Они должны принимать значения без паники.
	m.HttpRequestsTotal.WithLabelValues("GET", "/", "200").Inc()
	m.ActiveUsers.Inc()
}

func TestInitMetricsAndGetMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	InitMetrics(registry)

	m := GetMetrics()
	require.NotNil(t, m)
	require.NotNil(t, m.HttpRequestsTotal)

	// Повторная инициализация (sync.Once) не должна паниковать или менять состояние.
	InitMetrics(prometheus.NewRegistry())
	m2 := GetMetrics()
	require.Same(t, m, m2)

	m.HttpRequestsTotal.WithLabelValues("GET", "/x", "200").Inc()
	m.AuthValidateTokenTotal.WithLabelValues("ok").Inc()
	m.AuthValidateRefreshTokenTotal.WithLabelValues("ok").Inc()
	m.ActiveUsers.Set(5)
}
