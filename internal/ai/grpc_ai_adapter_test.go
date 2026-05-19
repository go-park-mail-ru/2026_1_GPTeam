package ai

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/groq"
	aiv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1"
	aimocks "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1/mocks"
)

func TestGrpcAiAdapter_Transcribe(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		setupMocks     func(m *aimocks.MockAiServiceClient)
		audioData      []byte
		filename       string
		expectedResult string
		expectedErr    bool
	}{
		{
			name: "successful transcription",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					Transcribe(gomock.Any(), &aiv1.TranscribeRequest{
						AudioData: []byte("audio-bytes"),
						Filename:  "voice.webm",
					}).
					Return(&aiv1.TranscribeResponse{Text: "Привет мир"}, nil)
			},
			audioData:      []byte("audio-bytes"),
			filename:       "voice.webm",
			expectedResult: "Привет мир",
			expectedErr:    false,
		},
		{
			name: "gRPC error",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					Transcribe(gomock.Any(), &aiv1.TranscribeRequest{
						AudioData: []byte("audio-bytes"),
						Filename:  "voice.webm",
					}).
					Return(nil, assert.AnError)
			},
			audioData:      []byte("audio-bytes"),
			filename:       "voice.webm",
			expectedResult: "",
			expectedErr:    true,
		},
		{
			name: "empty audio data",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					Transcribe(gomock.Any(), &aiv1.TranscribeRequest{
						AudioData: []byte{},
						Filename:  "empty.webm",
					}).
					Return(&aiv1.TranscribeResponse{Text: ""}, nil)
			},
			audioData:      []byte{},
			filename:       "empty.webm",
			expectedResult: "",
			expectedErr:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := aimocks.NewMockAiServiceClient(ctrl)
			c.setupMocks(mockClient)

			adapter := NewGrpcAiAdapter(mockClient)

			ctx := context.Background()
			result, err := adapter.Transcribe(ctx, c.audioData, c.filename)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expectedResult, result)
			}
		})
	}
}

func TestGrpcAiAdapter_ParseTransaction(t *testing.T) {
	t.Parallel()

	testDate := time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name          string
		setupMocks    func(m *aimocks.MockAiServiceClient)
		transcript    string
		types         []string
		categories    []string
		expectedDraft *groq.TransactionDraft
		expectedErr   bool
	}{
		{
			name: "successful parse transaction",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), &aiv1.ParseTransactionRequest{
						Transcript: "купил молоко за 150 рублей",
						Types:      []string{"expense"},
						Categories: []string{"food"},
					}).
					Return(&aiv1.ParseTransactionResponse{
						Draft: &aiv1.TransactionDraft{
							Value:       150.0,
							Type:        "expense",
							Category:    "food",
							Title:       "Молоко",
							Description: "Покупка молока",
							Date:        "2026-05-17",
						},
					}, nil)
			},
			transcript: "купил молоко за 150 рублей",
			types:      []string{"expense"},
			categories: []string{"food"},
			expectedDraft: &groq.TransactionDraft{
				RawText:     "купил молоко за 150 рублей",
				Value:       150.0,
				Type:        "expense",
				Category:    "food",
				Title:       "Молоко",
				Description: "Покупка молока",
				Date:        testDate,
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), &aiv1.ParseTransactionRequest{
						Transcript: "купил хлеб",
						Types:      []string{"expense"},
						Categories: []string{"food"},
					}).
					Return(nil, assert.AnError)
			},
			transcript:    "купил хлеб",
			types:         []string{"expense"},
			categories:    []string{"food"},
			expectedDraft: nil,
			expectedErr:   true,
		},
		{
			name: "nil draft response",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), &aiv1.ParseTransactionRequest{
						Transcript: "привет как дела",
						Types:      []string{"expense"},
						Categories: []string{"food"},
					}).
					Return(&aiv1.ParseTransactionResponse{Draft: nil}, nil)
			},
			transcript:    "привет как дела",
			types:         []string{"expense"},
			categories:    []string{"food"},
			expectedDraft: nil,
			expectedErr:   false,
		},
		{
			name: "empty date uses current time",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), &aiv1.ParseTransactionRequest{
						Transcript: "купил сыр",
						Types:      []string{"expense"},
						Categories: []string{"food"},
					}).
					Return(&aiv1.ParseTransactionResponse{
						Draft: &aiv1.TransactionDraft{
							Value:       200.0,
							Type:        "expense",
							Category:    "food",
							Title:       "Сыр",
							Description: "Покупка сыра",
							Date:        "",
						},
					}, nil)
			},
			transcript: "купил сыр",
			types:      []string{"expense"},
			categories: []string{"food"},
			expectedDraft: &groq.TransactionDraft{
				RawText:     "купил сыр",
				Value:       200.0,
				Type:        "expense",
				Category:    "food",
				Title:       "Сыр",
				Description: "Покупка сыра",
				Date:        time.Now(),
			},
			expectedErr: false,
		},
		{
			name: "invalid date format uses current time",
			setupMocks: func(m *aimocks.MockAiServiceClient) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), &aiv1.ParseTransactionRequest{
						Transcript: "купил воду",
						Types:      []string{"expense"},
						Categories: []string{"food"},
					}).
					Return(&aiv1.ParseTransactionResponse{
						Draft: &aiv1.TransactionDraft{
							Value:       50.0,
							Type:        "expense",
							Category:    "food",
							Title:       "Вода",
							Description: "Покупка воды",
							Date:        "invalid-date",
						},
					}, nil)
			},
			transcript: "купил воду",
			types:      []string{"expense"},
			categories: []string{"food"},
			expectedDraft: &groq.TransactionDraft{
				RawText:     "купил воду",
				Value:       50.0,
				Type:        "expense",
				Category:    "food",
				Title:       "Вода",
				Description: "Покупка воды",
				Date:        time.Now(),
			},
			expectedErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := aimocks.NewMockAiServiceClient(ctrl)
			c.setupMocks(mockClient)

			adapter := NewGrpcAiAdapter(mockClient)

			ctx := context.Background()
			draft, err := adapter.ParseTransaction(ctx, c.transcript, c.types, c.categories)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if c.expectedDraft == nil {
					assert.Nil(t, draft)
				} else {
					require.NotNil(t, draft)
					assert.Equal(t, c.expectedDraft.RawText, draft.RawText)
					assert.Equal(t, c.expectedDraft.Value, draft.Value)
					assert.Equal(t, c.expectedDraft.Type, draft.Type)
					assert.Equal(t, c.expectedDraft.Category, draft.Category)
					assert.Equal(t, c.expectedDraft.Title, draft.Title)
					assert.Equal(t, c.expectedDraft.Description, draft.Description)
					// For dates, we allow a small tolerance
					if !c.expectedDraft.Date.IsZero() {
						assert.True(t, draft.Date.Equal(c.expectedDraft.Date) ||
							draft.Date.Sub(c.expectedDraft.Date).Abs() < time.Second)
					}
				}
			}
		})
	}
}

func TestNewGrpcAiAdapter(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := aimocks.NewMockAiServiceClient(ctrl)
	adapter := NewGrpcAiAdapter(mockClient)

	require.NotNil(t, adapter)
	assert.Equal(t, mockClient, adapter.client)
}

func TestParseDate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		dateStr  string
		expected time.Time
	}{
		{
			name:     "valid date",
			dateStr:  "2026-05-17",
			expected: time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "empty string",
			dateStr:  "",
			expected: time.Now(),
		},
		{
			name:     "invalid format",
			dateStr:  "17-05-2026",
			expected: time.Now(),
		},
		{
			name:     "malformed date",
			dateStr:  "not-a-date",
			expected: time.Now(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			result := parseDate(c.dateStr)

			if c.dateStr == "" || c.dateStr == "17-05-2026" || c.dateStr == "not-a-date" {
				assert.True(t, result.Sub(c.expected).Abs() < time.Second)
			} else {
				assert.True(t, result.Equal(c.expected))
			}
		})
	}
}
