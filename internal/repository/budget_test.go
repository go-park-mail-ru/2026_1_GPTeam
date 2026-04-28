package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func newBudgetPostgres(t *testing.T) (*BudgetPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return &BudgetPostgres{db: mock}, mock
}

func TestBudgetPostgres_Create(t *testing.T) {
	t.Parallel()
	now := time.Now()
	budget := models.BudgetModel{
		Title:       "Trip",
		Description: "Paris",
		CreatedAt:   now,
		StartAt:     now,
		Target:      2000,
		Currency:    "RUB",
		Author:      1,
	}

	t.Run("success", func(t *testing.T) {
		repo, mock := newBudgetPostgres(t)
		mock.ExpectQuery(`insert into budget`).
			WithArgs(budget.Title, budget.Description, budget.CreatedAt, budget.StartAt, pgxmock.AnyArg(), budget.Actual, budget.Target, budget.Currency, budget.Author).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.Create(context.Background(), budget)
		require.NoError(t, err)
		require.Equal(t, 1, id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newBudgetPostgres(t)
		mock.ExpectQuery(`insert into budget`).
			WillReturnError(errors.New("db error"))

		_, err := repo.Create(context.Background(), budget)
		require.Error(t, err)
	})
}

func TestBudgetPostgres_GetById(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newBudgetPostgres(t)
		rows := pgxmock.NewRows([]string{"title", "description", "created_at", "start_at", "end_at", "updated_at", "actual", "target", "currency", "author", "active"}).
			AddRow("Title", "Desc", time.Now(), time.Now(), time.Now(), time.Now(), 0, 1000, "RUB", 1, true)

		mock.ExpectQuery(`select title, description, created_at, start_at, end_at, updated_at, actual, target, currency, author, active from budget`).
			WithArgs(1).
			WillReturnRows(rows)

		budget, err := repo.GetById(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, "Title", budget.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBudgetPostgres_GetIdsByUserId(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newBudgetPostgres(t)
		rows := pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)

		mock.ExpectQuery(`select id from budget`).
			WithArgs(1).
			WillReturnRows(rows)

		ids, err := repo.GetIdsByUserId(context.Background(), 1)
		require.NoError(t, err)
		require.Len(t, ids, 2)
	})

	t.Run("nothing found", func(t *testing.T) {
		repo, mock := newBudgetPostgres(t)
		mock.ExpectQuery(`select id from budget`).
			WithArgs(1).
			WillReturnRows(pgxmock.NewRows([]string{"id"}))

		_, err := repo.GetIdsByUserId(context.Background(), 1)
		require.ErrorIs(t, err, NothingInTableError)
	})
}

func TestBudgetPostgres_Delete(t *testing.T) {
	t.Parallel()

	repo, mock := newBudgetPostgres(t)
	mock.ExpectExec(`update budget set active = false`).
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}
