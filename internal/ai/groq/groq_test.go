package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroqClient_Transcribe(t *testing.T) {
	//t.Parallel()
	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))

	cases := []struct {
		name           string
		handler        http.HandlerFunc
		expectedResult string
		expectedErr    error
	}{
		{
			name: "success transcribe",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
				assert.Equal(t, fmt.Sprintf("Bearer %s", groqKey), r.Header.Get("Authorization"))

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(groqTranscriptResponse{Text: "Привет мир"})
			},
			expectedResult: "Привет мир",
			expectedErr:    nil,
		},
		{
			name: "empty result from groq",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(groqTranscriptResponse{Text: ""})
			},
			expectedResult: "",
			expectedErr:    ErrClientEmptyResult,
		},
		{
			name: "rate limit error 429",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				var errResp groqErrorResponse
				errResp.Error.Type = "rate_limit"
				errResp.Error.Code = "429"
				errResp.Error.Message = "Too many requests"
				_ = json.NewEncoder(w).Encode(errResp)
			},
			expectedResult: "",
			expectedErr:    ErrClientRateLimit,
		},
		{
			name: "bad request error 400",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				var errResp groqErrorResponse
				errResp.Error.Type = "invalid_file"
				_ = json.NewEncoder(w).Encode(errResp)
			},
			expectedResult: "",
			expectedErr:    ErrClientInvalidFile,
		},
		{
			name: "unauthorized error 401",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				var errResp groqErrorResponse
				errResp.Error.Type = "unauthorized"
				_ = json.NewEncoder(w).Encode(errResp)
			},
			expectedResult: "",
			expectedErr:    ErrClientUnauthorized,
		},
		{
			name: "internal server error 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("plain text error"))
			},
			expectedResult: "",
			expectedErr:    ErrInternalClient,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			//t.Parallel()

			server := httptest.NewServer(c.handler)
			defer server.Close()

			oldSTTURL := groqSTTURL
			groqSTTURL = server.URL
			defer func() { groqSTTURL = oldSTTURL }()

			client := NewGroqClient(groqKey, "")
			client.httpClient = &http.Client{Timeout: 60 * time.Second}

			ctx := context.Background()
			res, err := client.Transcribe(ctx, []byte("audio-bytes"), "voice.webm")

			if c.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, c.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expectedResult, res)
			}
		})
	}
}

func TestGroqClient_ParseTransaction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		transcript    string
		handler       http.HandlerFunc
		expectedDraft *TransactionDraft
		expectedErr   bool
	}{
		{
			name:          "empty transcript error",
			transcript:    "   ",
			handler:       func(w http.ResponseWriter, r *http.Request) {},
			expectedDraft: nil,
			expectedErr:   true,
		},
		{
			name:       "success parse transaction markdown block",
			transcript: "купил молоко за 150 рублей",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
				resp := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: "```json\n{\n  \"value\": 150.0,\n  \"type\": \"expense\",\n  \"category\": \"food\",\n  \"title\": \"Молоко\",\n  \"description\": \"Покупка молока\",\n  \"date\": \"2026-05-17\"\n}\n```",
							},
						},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedDraft: &TransactionDraft{
				Value:       150.0,
				Type:        "expense",
				Category:    "food",
				Title:       "Молоко",
				Description: "Покупка молока",
			},
			expectedErr: false,
		},
		{
			name:       "empty transaction output",
			transcript: "как дела",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: "{}",
							},
						},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedDraft: nil,
			expectedErr:   false,
		},
		{
			name:       "groq non-200 status code",
			transcript: "купил сыр",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedDraft: nil,
			expectedErr:   true,
		},
		{
			name:       "invalid inner JSON syntax",
			transcript: "купил хлеб",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: "{invalid-json}",
							},
						},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedDraft: nil,
			expectedErr:   true,
		},
	}

	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(c.handler)
			defer server.Close()

			oldChatURL := groqChatURL
			groqChatURL = server.URL
			defer func() { groqChatURL = oldChatURL }()

			client := NewGroqClient(groqKey, "")
			client.httpClient = &http.Client{Timeout: 60 * time.Second}

			ctx := context.Background()
			draft, err := client.ParseTransaction(ctx, c.transcript, []string{"expense"}, []string{"food"})

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if c.expectedDraft == nil {
					assert.Nil(t, draft)
				} else {
					require.NotNil(t, draft)
					assert.Equal(t, c.expectedDraft.Value, draft.Value)
					assert.Equal(t, c.expectedDraft.Type, draft.Type)
					assert.Equal(t, c.expectedDraft.Category, draft.Category)
					assert.Equal(t, c.expectedDraft.Title, draft.Title)
					assert.Equal(t, c.expectedDraft.Description, draft.Description)
				}
			}
		})
	}
}

func TestStripMarkdownFences(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "json code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "plain code block",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "no code block",
			input:    "{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "with whitespace",
			input:    "  ```json\n{\"key\": \"value\"}\n```  ",
			expected: "{\"key\": \"value\"}",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			result := stripMarkdownFences(c.input)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestNewGroqClient(t *testing.T) {
	t.Parallel()

	t.Run("client with empty proxy", func(t *testing.T) {
		client := NewGroqClient("test-key", "")
		require.NotNil(t, client)
		assert.Equal(t, "test-key", client.apiKey)
	})

	t.Run("client with valid proxy", func(t *testing.T) {
		client := NewGroqClient("test-key", "http://proxy.example.com:8080")
		require.NotNil(t, client)
		assert.Equal(t, "test-key", client.apiKey)
	})

	t.Run("client trims api key", func(t *testing.T) {
		client := NewGroqClient("  test-key  ", "")
		assert.Equal(t, "test-key", client.apiKey)
	})
}
