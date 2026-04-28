package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func newTransactionPostgres(t *testing.T) (*TransactionPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp),
	)
	require.NoError(t, err)
	return NewTransactionPostgres(mock), mock
}

func validTransactionModel(txDate time.Time) models.TransactionModel {
	return models.TransactionModel{
		Id:              42,
		UserId:          7,
		AccountId:       55,
		Value:           3850,
		Type:            "expense",
		Category:        "food",
		Title:           "Покупка продуктов",
		Description:     "Перекрёсток",
		TransactionDate: txDate,
	}
}

func validAccountModel(date time.Time) models.AccountModel {
	return models.AccountModel{
		Id:        55,
		Name:      "name",
		Balance:   100000,
		Currency:  "RUB",
		CreatedAt: date,
		UpdatedAt: date,
	}
}

func TestTransactionPostgres_Create(t *testing.T) {
	t.Parallel()
	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	account := validAccountModel(txDate)

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`insert into transaction`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(42))

		mock.ExpectExec(`update account set balance`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()

		id, err := repo.Create(context.Background(), tx, account)
		require.NoError(t, err)
		require.Equal(t, 42, id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique violation", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()
		// ИСПРАВЛЕНО: Добавлено WithArgs для совпадения с реальным запросом
		mock.ExpectQuery(`insert into transaction`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
		mock.ExpectRollback()

		_, err := repo.Create(context.Background(), tx, account)
		require.ErrorIs(t, err, TransactionDuplicatedDataError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionPostgres_GetIdsByUserId(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)
		rows := pgxmock.NewRows([]string{"id"}).AddRow(10).AddRow(11)

		mock.ExpectQuery(`select id from transaction`).
			WithArgs(7).
			WillReturnRows(rows)

		ids, err := repo.GetIdsByUserId(context.Background(), 7)
		require.NoError(t, err)
		require.Equal(t, []int{10, 11}, ids)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		// ИСПРАВЛЕНО: Имитируем успешный запрос, который вернул 0 строк
		mock.ExpectQuery(`select id from transaction`).
			WithArgs(7).
			WillReturnRows(pgxmock.NewRows([]string{"id"}))

		_, err := repo.GetIdsByUserId(context.Background(), 7)
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionPostgres_Update(t *testing.T) {
	t.Parallel()
	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	oldTx := validTransactionModel(txDate.Add(24 * time.Hour))
	account := validAccountModel(txDate)
	oldAccount := validAccountModel(txDate)

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()

		mock.ExpectQuery(`select value, type, account_id from transaction`).
			WithArgs(tx.Id, tx.UserId).
			WillReturnRows(pgxmock.NewRows([]string{"value", "type", "account_id"}).AddRow(float64(100), "expense", 55))

		mock.ExpectExec(`update transaction set`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectExec(`update account set balance`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()

		err := repo.Update(context.Background(), tx, oldTx, account, oldAccount)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`select value, type, account_id from transaction`).
			WithArgs(tx.Id, tx.UserId).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		err := repo.Update(context.Background(), tx, oldTx, account, oldAccount)
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionPostgres_Delete(t *testing.T) {
	t.Parallel()
	date := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	account := validAccountModel(date)

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`UPDATE transaction SET deleted_at`).
			WithArgs(42).
			WillReturnRows(pgxmock.NewRows([]string{"id", "type", "value", "account_id", "user_id", "category"}).
				AddRow(42, "EXPENSE", float64(100), 55, 7, "food"))

		mock.ExpectExec(`update account set balance`).
			WithArgs("EXPENSE", float64(100), 55).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectExec(`update budget set actual`).
			WithArgs("EXPENSE", float64(100), 7, "food").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()
		mock.ExpectRollback()

		id, err := repo.Delete(context.Background(), 42, account)

		require.NoError(t, err)
		require.Equal(t, 42, id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`UPDATE transaction SET deleted_at`).
			WithArgs(4).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()
		mock.ExpectRollback()

		_, err := repo.Delete(context.Background(), 4, account)
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionPostgres_Detail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)
		rows := pgxmock.NewRows([]string{"user_id", "account_id", "value", "type", "category", "title", "description", "created_at", "transaction_date", "updated_at"}).
			AddRow(7, 55, float64(3850), "expense", "food", "Покупка", "Описание", time.Now(), time.Now(), time.Now())

		mock.ExpectQuery(`select user_id, account_id`).
			WithArgs(42).
			WillReturnRows(rows)

		tx, err := repo.Detail(context.Background(), 42)
		require.NoError(t, err)
		require.Equal(t, 7, tx.UserId)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionPostgres_Search(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newTransactionPostgres(t)
		rows := pgxmock.NewRows([]string{"id", "user_id", "account_id", "value", "type", "category", "title", "description", "created_at", "transaction_date", "updated_at"}).
			AddRow(42, 7, 55, float64(3850), "expense", "food", "Покупка", "Описание", time.Now(), time.Now(), time.Now())

		mock.ExpectQuery(`select id, user_id`).
			WithArgs(7).
			WillReturnRows(rows)

		txs, err := repo.Search(context.Background(), 7, TransactionFilters{})
		require.NoError(t, err)
		require.Len(t, txs, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
