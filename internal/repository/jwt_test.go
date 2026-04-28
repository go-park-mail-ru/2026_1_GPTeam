package repository

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func newJwtPostgres(t *testing.T) (*JwtPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return NewJwtPostgres(mock), mock
}

func TestJwtPostgres_Create(t *testing.T) {
	t.Parallel()
	token := models.RefreshTokenModel{Uuid: "rt-1", UserId: 7, ExpiredAt: time.Now()}

	t.Run("success", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`insert into jwt`).
			WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.Create(context.Background(), token)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique violation", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`insert into jwt`).
			WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		err := repo.Create(context.Background(), token)
		require.ErrorIs(t, err, DuplicatedDataError)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("constraint error", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`insert into jwt`).
			WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

		err := repo.Create(context.Background(), token)
		require.ErrorIs(t, err, ConstraintError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestJwtPostgres_DeleteByUuid(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`delete from jwt`).
			WithArgs("rt-1").
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteByUuid(context.Background(), "rt-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`delete from jwt`).
			WithArgs("rt-1").
			WillReturnError(pgx.ErrNoRows)

		err := repo.DeleteByUuid(context.Background(), "rt-1")
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestJwtPostgres_DeleteByUserId(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`delete from jwt`).
			WithArgs(7).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteByUserId(context.Background(), 7)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectExec(`delete from jwt`).
			WithArgs(7).
			WillReturnError(pgx.ErrNoRows)

		err := repo.DeleteByUserId(context.Background(), 7)
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestJwtPostgres_Get(t *testing.T) {
	t.Parallel()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		rows := pgxmock.NewRows([]string{"user_id", "expired_at"}).AddRow(7, now)

		mock.ExpectQuery(`select user_id, expired_at from jwt`).
			WithArgs("rt-1").
			WillReturnRows(rows)

		token, err := repo.Get(context.Background(), "rt-1")
		require.NoError(t, err)
		require.Equal(t, 7, token.UserId)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newJwtPostgres(t)
		mock.ExpectQuery(`select user_id, expired_at from jwt`).
			WithArgs("rt-1").
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.Get(context.Background(), "rt-1")
		require.ErrorIs(t, err, NothingInTableError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
