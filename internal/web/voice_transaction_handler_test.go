package web

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func TestVoiceHandler_CreateVoiceTransaction(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{Id: 1}

	cases := []struct {
		name         string
		ctx          context.Context
		setupReq     func(w *multipart.Writer)
		setupMocks   func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase)
		expectedCode int
	}{
		{
			name: "не авторизован",
			ctx:  context.Background(),
			setupReq: func(w *multipart.Writer) {
				_, _ = w.CreateFormFile("audio", "test.wav")
			},
			setupMocks:   func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "отсутствует файл",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupReq: func(w *multipart.Writer) {
				_ = w.WriteField("other", "field")
			},
			setupMocks:   func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "успешная генерация черновика",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupReq: func(w *multipart.Writer) {
				part, _ := w.CreateFormFile("audio", "test.wav")
				_, _ = part.Write([]byte("fake audio content"))
			},
			setupMocks: func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase) {
				// ВАЖНО: Оставляем только те моки, которые реально вызываются в хендлере!
				enums.EXPECT().GetTransactionTypes().Return([]string{"expense"})
				enums.EXPECT().GetCategoryTypes().Return([]string{"food"})

				voiceSvc.EXPECT().CreateVoiceTransaction(gomock.Any(), gomock.Any(), "test.wav").
					Return(&models.TransactionDraft{
						Title:       "Обед",
						Description: "Поел в столовой",
						Value:       500,
						Type:        "expense",
						Category:    "food",
						Date:        time.Now(),
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "в речи не найдена транзакция",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupReq: func(w *multipart.Writer) {
				part, _ := w.CreateFormFile("audio", "test.wav")
				_, _ = part.Write([]byte("fake audio content"))
			},
			setupMocks: func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase) {
				voiceSvc.EXPECT().CreateVoiceTransaction(gomock.Any(), gomock.Any(), "test.wav").
					Return(&models.TransactionDraft{}, nil)
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "ошибка сервиса",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupReq: func(w *multipart.Writer) {
				part, _ := w.CreateFormFile("audio", "test.wav")
				_, _ = part.Write([]byte("fake audio content"))
			},
			setupMocks: func(voiceSvc *mocks.MockVoiceTransactionUseCase, enums *mocks.MockEnumsUseCase) {
				voiceSvc.EXPECT().CreateVoiceTransaction(gomock.Any(), gomock.Any(), "test.wav").
					Return(nil, errors.New("internal error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cases {
		c := c // Для параллельных тестов
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			voiceSvc := mocks.NewMockVoiceTransactionUseCase(ctrl)
			enums := mocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(voiceSvc, enums)

			handler := NewVoiceHandler(voiceSvc, enums)

			body := &bytes.Buffer{}
			mw := multipart.NewWriter(body)
			c.setupReq(mw)
			_ = mw.Close()

			req := httptest.NewRequest(http.MethodPost, "/transactions/voice", body).WithContext(c.ctx)
			req.Header.Set("Content-Type", mw.FormDataContentType())
			w := httptest.NewRecorder()

			handler.CreateVoiceTransaction(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
