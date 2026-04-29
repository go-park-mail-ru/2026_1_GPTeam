package application

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEnums_GetCategoryTypes(t *testing.T) {
	t.Run("get categories", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := mocks.NewMockEnumsRepository(ctrl)
		repo.EXPECT().GetCategoryTypesFromDB().Return([]string{"a", "b"})
		app := NewEnums(repo)
		categories := app.GetCategoryTypes()
		require.Equal(t, []string{"a", "b"}, categories)
	})
}

func TestEnums_GetCurrencyCodes(t *testing.T) {
	t.Run("get currencies", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := mocks.NewMockEnumsRepository(ctrl)
		repo.EXPECT().GetCurrencyCodesFromDB().Return([]string{"1", "2"})
		app := NewEnums(repo)
		currencies := app.GetCurrencyCodes()
		require.Equal(t, []string{"1", "2"}, currencies)
	})
}

func TestEnums_GetTransactionTypes(t *testing.T) {
	t.Run("get types", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := mocks.NewMockEnumsRepository(ctrl)
		repo.EXPECT().GetTransactionTypesFromDB().Return([]string{"a", "b"})
		app := NewEnums(repo)
		types := app.GetTransactionTypes()
		require.Equal(t, []string{"a", "b"}, types)
	})
}
