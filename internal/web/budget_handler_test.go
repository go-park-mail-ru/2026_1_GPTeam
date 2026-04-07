package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func TestBudgetHandler_GetBudgets(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(budgetApp *appmocks.MockBudgetUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().GetBudgetsOfUser(gomock.Any(), testUser).Return([]int{1, 2, 3}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "internal error",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().GetBudgetsOfUser(gomock.Any(), testUser).Return(nil, errors.New("db error"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			budgetApp := appmocks.NewMockBudgetUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(budgetApp)

			handler := NewBudgetHandler(budgetApp, enumsApp)
			req := httptest.NewRequest(http.MethodGet, "/budgets", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.GetBudgets(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestBudgetHandler_GetBudget(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}
	testBudget := models.BudgetModel{Id: 1, Title: "Test Budget"}

	cases := []struct {
		name         string
		id           string
		ctx          context.Context
		setupMocks   func(budgetApp *appmocks.MockBudgetUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			id:           "1",
			ctx:          context.Background(),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "empty id",
			id:           "",
			ctx:          context.WithValue(context.Background(), "user", testUser),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid id",
			id:           "abc",
			ctx:          context.WithValue(context.Background(), "user", testUser),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "not author",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().GetById(gomock.Any(), 1, testUser).Return(models.BudgetModel{}, application.UserNotAuthorOfBudgetError)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "not found",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().GetById(gomock.Any(), 1, testUser).Return(models.BudgetModel{}, repository.NothingInTableError)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "success",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().GetById(gomock.Any(), 1, testUser).Return(testBudget, nil)
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

			budgetApp := appmocks.NewMockBudgetUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(budgetApp)

			handler := NewBudgetHandler(budgetApp, enumsApp)
			req := httptest.NewRequest(http.MethodGet, "/budgets/", nil).WithContext(c.ctx)
			req.SetPathValue("id", c.id)

			w := httptest.NewRecorder()
			handler.GetBudget(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestBudgetHandler_Create(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}
	validBody := web_helpers.BudgetRequest{
		Title:       "New Budget",
		Description: "Desc",
		Target:      1000,
		Actual:      0,
		Currency:    "RUB",
		StartAt:     time.Now().Add(time.Hour),
		EndAt:       time.Now().Add(24 * time.Hour),
	}

	cases := []struct {
		name         string
		body         any
		ctx          context.Context
		setupMocks   func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			body:         validBody,
			ctx:          context.Background(),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "invalid body",
			body: "invalid json",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: web_helpers.BudgetRequest{Title: ""},
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase) {
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "duplicate data",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase) {
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				budgetApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(0, repository.DuplicatedDataError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "success",
			body: validBody,
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase, enumsApp *appmocks.MockEnumsUseCase) {
				enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB"})
				budgetApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
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

			budgetApp := appmocks.NewMockBudgetUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(budgetApp, enumsApp)

			handler := NewBudgetHandler(budgetApp, enumsApp)

			bodyBytes, _ := json.Marshal(c.body)
			if str, ok := c.body.(string); ok {
				bodyBytes = []byte(str)
			}

			req := httptest.NewRequest(http.MethodPost, "/budgets", bytes.NewReader(bodyBytes)).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.Create(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestBudgetHandler_Delete(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		id           string
		ctx          context.Context
		setupMocks   func(budgetApp *appmocks.MockBudgetUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			id:           "1",
			ctx:          context.Background(),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid id",
			id:           "abc",
			ctx:          context.WithValue(context.Background(), "user", testUser),
			setupMocks:   func(budgetApp *appmocks.MockBudgetUseCase) {},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "not found",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().Delete(gomock.Any(), 1, testUser).Return(repository.NothingInTableError)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "success",
			id:   "1",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(budgetApp *appmocks.MockBudgetUseCase) {
				budgetApp.EXPECT().Delete(gomock.Any(), 1, testUser).Return(nil)
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

			budgetApp := appmocks.NewMockBudgetUseCase(ctrl)
			enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
			c.setupMocks(budgetApp)

			handler := NewBudgetHandler(budgetApp, enumsApp)
			req := httptest.NewRequest(http.MethodDelete, "/budgets/", nil).WithContext(c.ctx)
			req.SetPathValue("id", c.id)

			w := httptest.NewRecorder()
			handler.Delete(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
