package logger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetLoggerWIthRequestId(t *testing.T) {
	cases := []struct {
		name        string
		ctx         context.Context
		setupLogger func()
	}{
		{
			name: "Logger is nil",
			ctx:  context.Background(),
			setupLogger: func() {
				mu.Lock()
				logger = nil
				mu.Unlock()
			},
		},
		{
			name: "Context without request_id",
			ctx:  context.Background(),
			setupLogger: func() {
				mu.Lock()
				logger = zap.NewNop()
				mu.Unlock()
			},
		},
		{
			name: "Context with valid request_id",
			ctx:  context.WithValue(context.Background(), "request_id", "12345"),
			setupLogger: func() {
				mu.Lock()
				logger = zap.NewNop()
				mu.Unlock()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setupLogger()
			log := GetLoggerWithRequestId(c.ctx)
			require.NotNil(t, log)
		})
	}
}

func TestModifyLoggerWithDBQuery(t *testing.T) {
	log := zap.NewNop()

	cases := []struct {
		name     string
		query    string
		args     []any
		duration time.Duration
	}{
		{
			name:     "Valid query with args",
			query:    "SELECT * FROM users WHERE id = $1",
			args:     []any{1},
			duration: time.Millisecond * 10,
		},
		{
			name:     "Query without args",
			query:    "SELECT 1",
			args:     nil,
			duration: time.Millisecond * 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			modified := ModifyLoggerWithDBQuery(log, c.query, c.args, c.duration)
			require.NotNil(t, modified)
		})
	}
}

func TestLoggerLifecycle(t *testing.T) {
	defer os.Remove("backend.log")

	err := InitLogger(false)
	require.NoError(t, err)

	log := GetLogger()
	require.NotNil(t, log)

	err = Close()
	require.NoError(t, err)

	logAfterClose := GetLogger()
	require.NotNil(t, logAfterClose)
}

func TestAccessLoggerLifecycle(t *testing.T) {
	defer os.Remove("access.log")

	err := InitAccessLogger()
	require.NoError(t, err)

	log := GetAccessLogger()
	require.NotNil(t, log)

	err = AccessClose()
	require.NoError(t, err)

	logAfterClose := GetAccessLogger()
	require.NotNil(t, logAfterClose)
}
