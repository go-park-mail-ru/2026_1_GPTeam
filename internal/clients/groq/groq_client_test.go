package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroqClient_Transcribe(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text": "купил кофе"}`))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "")

	require.NotNil(t, client)
}

func TestGroqClient_ParseTransaction(t *testing.T) {
	t.Parallel()

	client := NewGroqClient("test-key", "")

	cases := []struct {
		name        string
		transcript  string
		expectedErr bool
	}{
		{
			name:        "пустой текст",
			transcript:  "",
			expectedErr: true,
		},
		{
			name:        "валидный текст",
			transcript:  "купил кофе за 200 рублей",
			expectedErr: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			draft, err := client.ParseTransaction(context.Background(), c.transcript, []string{"e"}, []string{"c"}, []string{"RUB"})
			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NotNil(t, err)
			}
			require.Nil(t, draft)
		})
	}
}
