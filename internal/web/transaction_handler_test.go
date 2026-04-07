package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func TestTransactionHandler_GetTransactions(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(trxApp *appmocks.MockTransactionUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			setupMocks:   func(trxApp *appmocks.MockTransactionUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().GetTransactionIdsOfUser(gomock.Any(), testUser).Return([]int{1, 2}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "not found",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().GetTransactionIdsOfUser(gomock.Any(), testUser).Return(nil, repository.NothingInTableError)
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			trxApp := appmocks.NewMockTransactionUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			accApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(trxApp)

			handler := NewTransactionHandler(trxApp, enumsApp, accApp)
			req := httptest.NewRequest(http.MethodGet, "/transactions", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.Transactions(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestTransactionHandler_Create(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}
	validBody := web_helpers.TransactionRequest{
		AccountId:   1,
		Title:       "Test Trx",
		Description: "Desc",
		Value:       100,
		Type:        "income",
		Category:    "salary",
		Currency:    "RUB",
	}

	cases := []struct {
		name         string
		body         any
		ctx          context.Context
		setupMocks   func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase)
		expectedCode int
	}{
		{
			name: "unauthorized",
			body: validBody,
			ctx:  context.Background(),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "validation error",
			body: web_helpers.TransactionRequest{Title: ""},
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
				enumsApp.EXPECT().GetTransactionTypes().Return([]string{"income"})
				enumsApp.EXPECT().GetCategoryTypes().Return([]string{"salary"})
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "not author of account",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
				enumsApp.EXPECT().GetTransactionTypes().Return([]string{"income"})
				enumsApp.EXPECT().GetCategoryTypes().Return([]string{"salary"})
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				accApp.EXPECT().IsUserAuthorOfAccount(gomock.Any(), testUser.Id, 1).Return(false)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "success",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
				enumsApp.EXPECT().GetTransactionTypes().Return([]string{"income"})
				enumsApp.EXPECT().GetCategoryTypes().Return([]string{"salary"})
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				accApp.EXPECT().IsUserAuthorOfAccount(gomock.Any(), testUser.Id, 1).Return(true)
				trxApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(10, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			trxApp := appmocks.NewMockTransactionUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			accApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(trxApp, enumsApp, accApp)

			handler := NewTransactionHandler(trxApp, enumsApp, accApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(bodyBytes)).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.Transactions(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestTransactionHandler_Detail(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		id           string
		ctx          context.Context
		setupMocks   func(trxApp *appmocks.MockTransactionUseCase)
		expectedCode int
	}{
		{
			name:         "invalid id",
			id:           "abc",
			ctx:          context.WithValue(context.Background(), "user", testUser),
			setupMocks:   func(trxApp *appmocks.MockTransactionUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "forbidden",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().Detail(gomock.Any(), 1, testUser.Id).Return(models.TransactionModel{}, application.ForbiddenError)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "success",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().Detail(gomock.Any(), 1, testUser.Id).Return(models.TransactionModel{Id: 1}, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			trxApp := appmocks.NewMockTransactionUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			accApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(trxApp)

			handler := NewTransactionHandler(trxApp, enumsApp, accApp)
			req := httptest.NewRequest(http.MethodGet, "/transactions/", nil).WithContext(c.ctx)
			req.SetPathValue("id", c.id)

			w := httptest.NewRecorder()
			handler.Transaction(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestTransactionHandler_Update(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}
	validBody := web_helpers.TransactionRequest{
		AccountId:   1,
		Title:       "Test Trx",
		Description: "Desc",
		Value:       100,
		Type:        "income",
		Category:    "salary",
		Currency:    "RUB",
	}

	cases := []struct {
		name         string
		id           string
		body         any
		ctx          context.Context
		setupMocks   func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase)
		expectedCode int
	}{
		{
			name: "invalid id",
			id:   "abc",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "success",
			id:   "1",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase, enumsApp *appmocks.MockEnumsUseCase, accApp *appmocks.MockAccountUseCase) {
				enumsApp.EXPECT().GetTransactionTypes().Return([]string{"income"})
				enumsApp.EXPECT().GetCategoryTypes().Return([]string{"salary"})
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				accApp.EXPECT().IsUserAuthorOfAccount(gomock.Any(), testUser.Id, 1).Return(true)
				trxApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			trxApp := appmocks.NewMockTransactionUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			accApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(trxApp, enumsApp, accApp)

			handler := NewTransactionHandler(trxApp, enumsApp, accApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPut, "/transactions/", bytes.NewReader(bodyBytes)).WithContext(c.ctx)
			req.SetPathValue("id", c.id)

			w := httptest.NewRecorder()
			handler.Transaction(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestTransactionHandler_Delete(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		id           string
		ctx          context.Context
		setupMocks   func(trxApp *appmocks.MockTransactionUseCase)
		expectedCode int
	}{
		{
			name:         "invalid id",
			id:           "abc",
			ctx:          context.WithValue(context.Background(), "user", testUser),
			setupMocks:   func(trxApp *appmocks.MockTransactionUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().Delete(gomock.Any(), 1, testUser.Id).Return(0, repository.NothingInTableError)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "success",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(trxApp *appmocks.MockTransactionUseCase) {
				trxApp.EXPECT().Delete(gomock.Any(), 1, testUser.Id).Return(1, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			trxApp := appmocks.NewMockTransactionUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			accApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(trxApp)

			handler := NewTransactionHandler(trxApp, enumsApp, accApp)
			req := httptest.NewRequest(http.MethodDelete, "/transactions/", nil).WithContext(c.ctx)
			req.SetPathValue("id", c.id)

			w := httptest.NewRecorder()
			handler.Transaction(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
