package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func TestVoiceTransactionService_CreateVoiceTransaction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMocks  func(client *mocks.MockAIConsultantClient, enums *mocks.MockEnumsUseCase)
		expectedErr bool
	}{
		{
			name: "успешная обработка",
			setupMocks: func(client *mocks.MockAIConsultantClient, enums *mocks.MockEnumsUseCase) {
				enums.EXPECT().GetTransactionTypes().Return([]string{"expense"})
				enums.EXPECT().GetCategoryTypes().Return([]string{"food"})
				enums.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				client.EXPECT().Transcribe(gomock.Any(), gomock.Any(), "test.wav").Return("купил хлеб", nil)
				client.EXPECT().ParseTransaction(gomock.Any(), "купил хлеб", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.TransactionDraft{Title: "Хлеб"}, nil)
			},
			expectedErr: false,
		},
		{
			name: "ошибка транскрипции",
			setupMocks: func(client *mocks.MockAIConsultantClient, enums *mocks.MockEnumsUseCase) {
				client.EXPECT().Transcribe(gomock.Any(), gomock.Any(), "test.wav").Return("", errors.New("stt error"))
			},
			expectedErr: true,
		},
		{
			name: "ошибка парсинга",
			setupMocks: func(client *mocks.MockAIConsultantClient, enums *mocks.MockEnumsUseCase) {
				enums.EXPECT().GetTransactionTypes().Return([]string{"expense"})
				enums.EXPECT().GetCategoryTypes().Return([]string{"food"})
				enums.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				client.EXPECT().Transcribe(gomock.Any(), gomock.Any(), "test.wav").Return("купил хлеб", nil)
				client.EXPECT().ParseTransaction(gomock.Any(), "купил хлеб", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("parse error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := mocks.NewMockAIConsultantClient(ctrl)
			enums := mocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(client, enums)

			svc := NewVoiceTransactionService(client, enums)
			draft, err := svc.CreateVoiceTransaction(context.Background(), []byte("audio"), "test.wav")

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, draft)
			} else {
				require.NoError(t, err)
				require.NotNil(t, draft)
			}
		})
	}
}
