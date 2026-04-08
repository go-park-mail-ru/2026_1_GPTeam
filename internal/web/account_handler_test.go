package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func ctxWithUser(user models.UserModel) context.Context {
	return context.WithValue(context.Background(), "user", user)
}

func TestAccountHandler_GetAccount(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(accountApp *appmocks.MockAccountUseCase)
		expectedCode int
		checkBody    func(t *testing.T, body map[string]any)
	}{
		{
			name: "нет пользователя в контексте → 401",
			ctx:  context.Background(),
			setupMocks: func(accountApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "аккаунт не найден → 404",
			ctx:  ctxWithUser(testUser),
			setupMocks: func(accountApp *appmocks.MockAccountUseCase) {
				accountApp.EXPECT().GetAccountIdByUserId(gomock.Any(), testUser.Id).Return(0, application.ErrAccountNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "внутренняя ошибка → 500",
			ctx:  ctxWithUser(testUser),
			setupMocks: func(accountApp *appmocks.MockAccountUseCase) {
				accountApp.EXPECT().GetAccountIdByUserId(gomock.Any(), testUser.Id).Return(0, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "успешно → 200 с account_id",
			ctx:  ctxWithUser(testUser),
			setupMocks: func(accountApp *appmocks.MockAccountUseCase) {
				accountApp.EXPECT().GetAccountIdByUserId(gomock.Any(), testUser.Id).Return(42, nil)
			},
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]any) {
				require.EqualValues(t, 42, body["account_id"])
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(accountApp)

			handler := NewAccountHandler(accountApp)
			req := httptest.NewRequest(http.MethodGet, "/account", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.GetAccount(w, req)

			require.Equal(t, c.expectedCode, w.Code)

			if c.checkBody != nil {
				var body map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
				c.checkBody(t, body)
			}
		})
	}
}
