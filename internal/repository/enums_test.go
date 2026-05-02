package repository

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestNewEnumsPostgres(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()

	mock.ExpectQuery(`select enumlabel from pg_enum where enumtypid = 'currency_code'`).
		WillReturnRows(pgxmock.NewRows([]string{"enumlabel"}).AddRow("RUB").AddRow("USD"))

	mock.ExpectQuery(`select enumlabel from pg_enum where enumtypid = 'transaction_type'`).
		WillReturnRows(pgxmock.NewRows([]string{"enumlabel"}).AddRow("INCOME").AddRow("EXPENSE"))

	mock.ExpectQuery(`select enumlabel from pg_enum where enumtypid = 'category_type'`).
		WillReturnRows(pgxmock.NewRows([]string{"enumlabel"}).AddRow("food").AddRow("rent"))

	repo, err := NewEnumsPostgres(ctx, mock)
	require.NoError(t, err)

	require.Equal(t, []string{"RUB", "USD"}, repo.GetCurrencyCodesFromDB())
	require.Equal(t, []string{"INCOME", "EXPENSE"}, repo.GetTransactionTypesFromDB())
	require.Equal(t, []string{"food", "rent"}, repo.GetCategoryTypesFromDB())

	require.NoError(t, mock.ExpectationsWereMet())
}
