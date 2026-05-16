package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/groq"
	aimocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/mocks"
	aiv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1"
)

func TestServer_Transcribe(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		req          *aiv1.TranscribeRequest
		setupMocks   func(m *aimocks.MockAiService)
		expectedResp *aiv1.TranscribeResponse
		expectedCode codes.Code
	}{
		{
			name: "missing audio data",
			req: &aiv1.TranscribeRequest{
				AudioData: nil,
				Filename:  "test.webm",
			},
			setupMocks:   func(m *aimocks.MockAiService) {},
			expectedResp: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "ai service layer failure",
			req: &aiv1.TranscribeRequest{
				AudioData: []byte("some-audio-content"),
				Filename:  "voice.webm",
			},
			setupMocks: func(m *aimocks.MockAiService) {
				m.EXPECT().
					Transcribe(gomock.Any(), []byte("some-audio-content"), "voice.webm").
					Return("", errors.New("groq down"))
			},
			expectedResp: nil,
			expectedCode: codes.Internal,
		},
		{
			name: "successful transcription",
			req: &aiv1.TranscribeRequest{
				AudioData: []byte("raw-bytes"),
				Filename:  "audio.ogg",
			},
			setupMocks: func(m *aimocks.MockAiService) {
				m.EXPECT().
					Transcribe(gomock.Any(), []byte("raw-bytes"), "audio.ogg").
					Return("Текст из аудио", nil)
			},
			expectedResp: &aiv1.TranscribeResponse{
				Text: "Текст из аудио",
			},
			expectedCode: codes.OK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAI := aimocks.NewMockAiService(ctrl)
			c.setupMocks(mockAI)

			server := &Server{
				AI: mockAI,
			}

			resp, err := server.Transcribe(context.Background(), c.req)

			if c.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, c.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expectedResp.Text, resp.Text)
			}
		})
	}
}

func TestServer_ParseTransaction(t *testing.T) {
	t.Parallel()

	testDate, err := time.Parse("2006-01-02", "2026-05-17")
	require.NoError(t, err)

	cases := []struct {
		name         string
		req          *aiv1.ParseTransactionRequest
		setupMocks   func(m *aimocks.MockAiService)
		expectedResp *aiv1.ParseTransactionResponse
		expectedCode codes.Code
	}{
		{
			name: "missing transcript",
			req: &aiv1.ParseTransactionRequest{
				Transcript: "",
			},
			setupMocks:   func(m *aimocks.MockAiService) {},
			expectedResp: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "ai service core crash",
			req: &aiv1.ParseTransactionRequest{
				Transcript: "Расход 500 рублей",
				Types:      []string{"expense"},
				Categories: []string{"food"},
			},
			setupMocks: func(m *aimocks.MockAiService) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), "Расход 500 рублей", []string{"expense"}, []string{"food"}).
					Return(nil, errors.New("llm parse panic"))
			},
			expectedResp: nil,
			expectedCode: codes.Internal,
		},
		{
			name: "transaction not found / empty result",
			req: &aiv1.ParseTransactionRequest{
				Transcript: "Привет, как твои дела?",
			},
			setupMocks: func(m *aimocks.MockAiService) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), "Привет, как твои дела?", gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedResp: nil,
			expectedCode: codes.NotFound,
		},
		{
			name: "successful extraction mapping",
			req: &aiv1.ParseTransactionRequest{
				Transcript: "Потратил 300 на такси",
				Types:      []string{"expense"},
				Categories: []string{"taxi"},
			},
			setupMocks: func(m *aimocks.MockAiService) {
				m.EXPECT().
					ParseTransaction(gomock.Any(), "Потратил 300 на такси", []string{"expense"}, []string{"taxi"}).
					Return(&groq.TransactionDraft{
						Value:       300,
						Type:        "expense",
						Category:    "taxi",
						Title:       "Такси",
						Description: "Поездка до дома",
						Date:        testDate,
					}, nil)
			},
			expectedResp: &aiv1.ParseTransactionResponse{
				Draft: &aiv1.TransactionDraft{
					Value:       300,
					Type:        "expense",
					Category:    "taxi",
					Title:       "Такси",
					Description: "Поездка до дома",
					Date:        "2026-05-17",
				},
			},
			expectedCode: codes.OK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAI := aimocks.NewMockAiService(ctrl)
			c.setupMocks(mockAI)

			server := &Server{
				AI: mockAI,
			}

			resp, err := server.ParseTransaction(context.Background(), c.req)

			if c.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, c.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, c.expectedResp.Draft.Value, resp.Draft.Value)
				assert.Equal(t, c.expectedResp.Draft.Type, resp.Draft.Type)
				assert.Equal(t, c.expectedResp.Draft.Category, resp.Draft.Category)
				assert.Equal(t, c.expectedResp.Draft.Title, resp.Draft.Title)
				assert.Equal(t, c.expectedResp.Draft.Description, resp.Draft.Description)
				assert.Equal(t, c.expectedResp.Draft.Date, resp.Draft.Date)
			}
		})
	}
}
