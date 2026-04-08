package context_helper_test

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/stretchr/testify/require"
)

func TestGetRequestIdFromContext(t *testing.T) {
	cases := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "Empty context",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "Invalid type in context",
			ctx:      context.WithValue(context.Background(), "request_id", 12345),
			expected: "",
		},
		{
			name:     "Valid request_id",
			ctx:      context.WithValue(context.Background(), "request_id", "test-req-id"),
			expected: "test-req-id",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			result := context_helper.GetRequestIdFromContext(c.ctx)
			require.Equal(t, c.expected, result)
		})
	}
}
