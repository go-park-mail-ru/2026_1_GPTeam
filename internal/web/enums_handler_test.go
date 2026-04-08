package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
)

func TestEnumsHandler_CurrencyCodes(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
	enumsApp.EXPECT().GetCurrencyCodes().Return([]string{"RUB", "USD"})

	handler := NewEnumsHandler(enumsApp)
	req := httptest.NewRequest(http.MethodGet, "/enums/currencies", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test"))

	w := httptest.NewRecorder()
	handler.CurrencyCodes(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "RUB")
}

func TestEnumsHandler_TransactionTypes(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
	enumsApp.EXPECT().GetTransactionTypes().Return([]string{"income", "expense"})

	handler := NewEnumsHandler(enumsApp)
	req := httptest.NewRequest(http.MethodGet, "/enums/transaction-types", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test"))

	w := httptest.NewRecorder()
	handler.TransactionTypes(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "income")
}

func TestEnumsHandler_CategoryTypes(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	enumsApp := appmocks.NewMockEnumsUseCase(ctrl)
	enumsApp.EXPECT().GetCategoryTypes().Return([]string{"food", "transport"})

	handler := NewEnumsHandler(enumsApp)
	req := httptest.NewRequest(http.MethodGet, "/enums/categories", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test"))

	w := httptest.NewRecorder()
	handler.CategoryTypes(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "food")
}
