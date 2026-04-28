package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockTransport перехватывает HTTP-запросы и возвращает замоканые ответы
type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func TestGroqClient_Transcribe(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		mockHandler   func(req *http.Request) (*http.Response, error)
		audioData     []byte
		filename      string
		expectedText  string
		expectedError bool
	}{
		{
			name: "success",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))
				require.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"text": "купил кофе за 200 рублей"}`)),
					Header:     make(http.Header),
				}, nil
			},
			audioData:     []byte("fake audio data"),
			filename:      "test.webm",
			expectedText:  "купил кофе за 200 рублей",
			expectedError: false,
		},
		{
			name: "rate limit error",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body: io.NopCloser(strings.NewReader(
						`{"error":{"message":"rate limit","type":"limits","code":"rate_limit_exceeded"}}`,
					)),
					Header: make(http.Header),
				}, nil
			},
			audioData:     []byte("fake audio data"),
			filename:      "test.webm",
			expectedText:  "",
			expectedError: true,
		},
		{
			name: "empty transcript",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"text": ""}`)),
					Header:     make(http.Header),
				}, nil
			},
			audioData:     []byte("fake audio data"),
			filename:      "test.webm",
			expectedText:  "",
			expectedError: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			client := &GroqClient{
				apiKey: "test-key",
				httpClient: &http.Client{
					Transport: &mockTransport{handler: c.mockHandler},
				},
			}

			text, err := client.Transcribe(context.Background(), c.audioData, c.filename)
			if c.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedText, text)
			}
		})
	}
}

func TestGroqClient_ParseTransaction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		mockHandler    func(req *http.Request) (*http.Response, error)
		transcript     string
		types          []string
		categories     []string
		expectedErr    bool
		expectNilDraft bool
	}{
		{
			name: "empty transcript",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				// Не должно вызываться — проверка на пустой текст происходит до запроса
				return nil, nil
			},
			transcript:     "",
			types:          []string{"expense", "income"},
			categories:     []string{"food", "transport"},
			expectedErr:    true,
			expectNilDraft: true,
		},
		{
			name: "success with valid json",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))
				require.Equal(t, "application/json", req.Header.Get("Content-Type"))

				var body groqChatRequest
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					return nil, err
				}
				require.Equal(t, "llama-3.3-70b-versatile", body.Model)
				require.Len(t, body.Messages, 2)
				require.Equal(t, "system", body.Messages[0].Role)
				require.Equal(t, "user", body.Messages[1].Role)

				response := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: `{"value":200,"type":"expense","category":"food","title":"Кофе","description":"Покупка кофе в кафе","date":"2024-01-15"}`,
							},
						},
					},
				}
				respBody, _ := json.Marshal(response)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     make(http.Header),
				}, nil
			},
			transcript:     "купил кофе за 200 рублей",
			types:          []string{"expense", "income"},
			categories:     []string{"food", "transport"},
			expectedErr:    false,
			expectNilDraft: false,
		},
		{
			name: "empty json response - no transaction",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				response := groqChatResponse{
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
				respBody, _ := json.Marshal(response)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     make(http.Header),
				}, nil
			},
			transcript:     "привет как дела",
			types:          []string{"expense", "income"},
			categories:     []string{"food", "transport"},
			expectedErr:    false,
			expectNilDraft: true,
		},
		{
			name: "json with markdown fences",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				response := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: "```json\n{\"value\":150.5,\"type\":\"income\",\"category\":\"salary\",\"title\":\"Зарплата\",\"description\":\"Поступление зарплаты\",\"date\":\"2024-01-20\"}\n```",
							},
						},
					},
				}
				respBody, _ := json.Marshal(response)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     make(http.Header),
				}, nil
			},
			transcript:     "получил зарплату 150 с половиной",
			types:          []string{"expense", "income"},
			categories:     []string{"food", "salary"},
			expectedErr:    false,
			expectNilDraft: false,
		},
		{
			name: "invalid json from llm",
			mockHandler: func(req *http.Request) (*http.Response, error) {
				response := groqChatResponse{
					Choices: []struct {
						Message chatMessage `json:"message"`
					}{
						{
							Message: chatMessage{
								Role:    "assistant",
								Content: `{"value":"not-a-number","type":"expense"}`,
							},
						},
					},
				}
				respBody, _ := json.Marshal(response)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     make(http.Header),
				}, nil
			},
			transcript:     "странная фраза",
			types:          []string{"expense"},
			categories:     []string{"food"},
			expectedErr:    true,
			expectNilDraft: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			client := &GroqClient{
				apiKey: "test-key",
				httpClient: &http.Client{
					Transport: &mockTransport{handler: c.mockHandler},
				},
			}

			draft, err := client.ParseTransaction(context.Background(), c.transcript, c.types, c.categories)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if c.expectNilDraft {
				require.Nil(t, draft)
			} else {
				require.NotNil(t, draft)
				// Базовая валидация структуры
				require.Equal(t, c.transcript, draft.RawText)
				require.NotEmpty(t, draft.Title)
				require.NotEmpty(t, draft.Description)
			}
		})
	}
}
