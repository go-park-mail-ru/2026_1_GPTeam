package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTransactionRequest(t *testing.T, method, target string, body any, user *models.UserModel) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")

	if user != nil {
		ctx := context.WithValue(req.Context(), "user", *user)
		req = req.WithContext(ctx)
	}

	return req
}

func TestTransactionHandler_Create_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	reqBody := map[string]any{
		"account_id":       55,
		"value":            3850.0,
		"type":             "expense",
		"category":         "food",
		"currency":         "RUB",
		"title":            "Покупка продуктов",
		"description":      "Перекрёсток",
		"transaction_date": txDate.Format(time.RFC3339),
	}
	user := models.UserModel{Id: 7, Username: "mike"}

	enumsUC.EXPECT().GetTransactionTypes().Return([]string{"income", "expense"})
	enumsUC.EXPECT().GetCategoryTypes().Return([]string{"food", "salary"})
	enumsUC.EXPECT().GetCurrencyCodes().Return([]string{"RUB", "USD"})
	accountUC.EXPECT().IsUserAuthorOfAccount(gomock.Any(), 7, 55).Return(true)
	txUC.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.TransactionModel{})).DoAndReturn(
		func(_ context.Context, got models.TransactionModel) (int, error) {
			require.Equal(t, 7, got.UserId)
			require.Equal(t, 55, got.AccountId)
			require.Equal(t, 3850.0, got.Value)
			require.Equal(t, "expense", got.Type)
			require.Equal(t, "food", got.Category)
			require.Equal(t, "RUB", got.Currency)
			require.Equal(t, "Покупка продуктов", got.Title)
			require.Equal(t, "Перекрёсток", got.Description)
			require.Equal(t, txDate, got.TransactionDate)
			require.False(t, got.CreatedAt.IsZero())
			return 101, nil
		},
	)

	req := newTransactionRequest(t, http.MethodPost, "/transactions", reqBody, &user)
	rr := httptest.NewRecorder()

	handler.Transactions(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code          int `json:"code"`
		TransactionID int `json:"transaction_id"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, 101, resp.TransactionID)
}

func TestTransactionHandler_Create_ForbiddenWhenAccountNotOwned(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	reqBody := map[string]any{
		"account_id":       55,
		"value":            100.0,
		"type":             "expense",
		"category":         "food",
		"currency":         "RUB",
		"title":            "Обед",
		"description":      "Кафе",
		"transaction_date": txDate.Format(time.RFC3339),
	}

	enumsUC.EXPECT().GetTransactionTypes().Return([]string{"income", "expense"})
	enumsUC.EXPECT().GetCategoryTypes().Return([]string{"food", "salary"})
	enumsUC.EXPECT().GetCurrencyCodes().Return([]string{"RUB", "USD"})
	accountUC.EXPECT().IsUserAuthorOfAccount(gomock.Any(), 7, 55).Return(false)

	req := newTransactionRequest(t, http.MethodPost, "/transactions", reqBody, &models.UserModel{Id: 7})
	rr := httptest.NewRecorder()

	handler.Transactions(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestTransactionHandler_GetTransactions_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txUC.EXPECT().GetTransactionIdsOfUser(gomock.Any(), gomock.AssignableToTypeOf(models.UserModel{})).DoAndReturn(
		func(_ context.Context, user models.UserModel) ([]int, error) {
			require.Equal(t, 7, user.Id)
			return []int{10, 20, 30}, nil
		},
	)

	req := newTransactionRequest(t, http.MethodGet, "/transactions", nil, &models.UserModel{Id: 7})
	rr := httptest.NewRecorder()

	handler.Transactions(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int   `json:"code"`
		Len  int   `json:"len"`
		IDs  []int `json:"ids"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 3, resp.Len)
	require.Equal(t, []int{10, 20, 30}, resp.IDs)
}

func TestTransactionHandler_Update_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	reqBody := map[string]any{
		"account_id":       55,
		"value":            100.0,
		"type":             "expense",
		"category":         "food",
		"currency":         "RUB",
		"title":            "Обед",
		"description":      "Кафе",
		"transaction_date": txDate.Format(time.RFC3339),
	}

	enumsUC.EXPECT().GetTransactionTypes().Return([]string{"income", "expense"})
	enumsUC.EXPECT().GetCategoryTypes().Return([]string{"food", "salary"})
	enumsUC.EXPECT().GetCurrencyCodes().Return([]string{"RUB", "USD"})
	accountUC.EXPECT().IsUserAuthorOfAccount(gomock.Any(), 7, 55).Return(true)
	txUC.EXPECT().Update(gomock.Any(), gomock.AssignableToTypeOf(models.TransactionModel{})).Return(repository.NothingInTableError)

	req := newTransactionRequest(t, http.MethodPut, "/transactions/42", reqBody, &models.UserModel{Id: 7})
	req.SetPathValue("id", "42")
	rr := httptest.NewRecorder()

	handler.Transaction(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestTransactionHandler_Delete_Forbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txUC.EXPECT().Delete(gomock.Any(), 42, 7).Return(0, application.ForbiddenError)

	req := newTransactionRequest(t, http.MethodDelete, "/transactions/42", nil, &models.UserModel{Id: 7})
	req.SetPathValue("id", "42")
	rr := httptest.NewRecorder()

	handler.Transaction(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestTransactionHandler_Detail_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	txUC := appmocks.NewMockTransactionUseCase(ctrl)
	enumsUC := appmocks.NewMockEnumsUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewTransactionHandler(txUC, enumsUC, accountUC)

	txDate := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	txUC.EXPECT().Detail(gomock.Any(), 42, 7).Return(models.TransactionModel{
		Id:              42,
		UserId:          7,
		AccountId:       55,
		Value:           3850,
		Type:            "expense",
		Category:        "food",
		Currency:        "RUB",
		Title:           "Покупка продуктов",
		Description:     "Перекрёсток",
		CreatedAt:       createdAt,
		TransactionDate: txDate,
	}, nil)

	req := newTransactionRequest(t, http.MethodGet, "/transactions/42", nil, &models.UserModel{Id: 7})
	req.SetPathValue("id", "42")
	rr := httptest.NewRecorder()

	handler.Transaction(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code        int `json:"code"`
		Transaction struct {
			ID        int     `json:"id"`
			UserID    int     `json:"user_id"`
			AccountID int     `json:"account_id"`
			Value     float64 `json:"value"`
			Title     string  `json:"title"`
		} `json:"transaction"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 42, resp.Transaction.ID)
	require.Equal(t, 7, resp.Transaction.UserID)
	require.Equal(t, 55, resp.Transaction.AccountID)
	require.Equal(t, 3850.0, resp.Transaction.Value)
	require.Equal(t, "Покупка продуктов", resp.Transaction.Title)
}
