package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func newSupportPostgres(t *testing.T) (*SupportPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp),
	)
	require.NoError(t, err)
	return NewPostgresSupport(mock), mock
}

func TestSupportPostgres_Create(t *testing.T) {
	now := time.Now()
	model := models.SupportModel{
		Id:        0,
		UserId:    1,
		Category:  "test",
		Message:   "some text",
		Status:    "OPEN",
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}
	wrongModel := model
	wrongModel.Message = strings.Repeat("a", 260)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("insert into support").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
		id, err := repository.Create(context.Background(), model)
		require.NoError(t, err)
		require.Equal(t, 1, id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("insert into support").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(ConstraintError)
		id, err := repository.Create(context.Background(), wrongModel)
		require.ErrorIs(t, err, ConstraintError)
		require.Equal(t, -1, id)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSupportPostgres_GetById(t *testing.T) {
	now := time.Now()
	expected := models.SupportModel{
		Id:        1,
		UserId:    1,
		Category:  "test",
		Message:   "some text",
		Status:    "OPEN",
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select user_id, category, message, status, created_at, updated_at from support").
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"user_id", "category", "message", "status", "created_at", "updated_at"}).AddRow(1, "test", "some text", "OPEN", now, now))
		support, err := repository.GetById(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, expected, support)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select user_id, category, message, status, created_at, updated_at from support").
			WithArgs(pgxmock.AnyArg()).
			WillReturnError(pgx.ErrNoRows)
		support, err := repository.GetById(context.Background(), 1)
		require.ErrorIs(t, err, NothingInTableError)
		require.Equal(t, models.SupportModel{}, support)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSupportPostgres_GetAll(t *testing.T) {
	now := time.Now()
	expectedSupports := []models.SupportModel{
		{
			Id:        1,
			UserId:    5,
			Category:  "a",
			Message:   "text",
			Status:    "OPEN",
			CreatedAt: now,
			UpdatedAt: now,
			Deleted:   false,
		},
		{
			Id:        2,
			UserId:    5,
			Category:  "b",
			Message:   "text",
			Status:    "CLOSED",
			CreatedAt: now,
			UpdatedAt: now.Add(24 * time.Hour),
			Deleted:   false,
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select id, user_id, category, message, status, created_at, updated_at from support").
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "category", "message", "status", "created_at", "updated_at"}).
				AddRow(1, 5, "a", "text", "OPEN", now, now).
				AddRow(2, 5, "b", "text", "CLOSED", now, now.Add(24*time.Hour)),
			)
		supports, err := repository.GetAll(context.Background())
		require.NoError(t, err)
		require.Equal(t, expectedSupports, supports)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var emptySupports []models.SupportModel
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select id, user_id, category, message, status, created_at, updated_at from support").
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "category", "message", "status", "created_at", "updated_at"}))
		supports, err := repository.GetAll(context.Background())
		require.NoError(t, err)
		require.Equal(t, emptySupports, supports)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSupportPostgres_GetAllByUser(t *testing.T) {
	now := time.Now()
	expectedSupports := []models.SupportModel{
		{
			Id:        1,
			UserId:    5,
			Category:  "a",
			Message:   "text",
			Status:    "OPEN",
			CreatedAt: now,
			UpdatedAt: now,
			Deleted:   false,
		},
		{
			Id:        2,
			UserId:    5,
			Category:  "b",
			Message:   "text",
			Status:    "CLOSED",
			CreatedAt: now,
			UpdatedAt: now.Add(24 * time.Hour),
			Deleted:   false,
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select id, category, message, status, created_at, updated_at from support").
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "category", "message", "status", "created_at", "updated_at"}).
				AddRow(1, "a", "text", "OPEN", now, now).
				AddRow(2, "b", "text", "CLOSED", now, now.Add(24*time.Hour)),
			)
		supports, err := repository.GetAllByUser(context.Background(), 5)
		require.NoError(t, err)
		require.Equal(t, expectedSupports, supports)
		require.Equal(t, 5, supports[0].UserId)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other user", func(t *testing.T) {
		t.Parallel()
		var emptySupports []models.SupportModel
		repository, mock := newSupportPostgres(t)
		mock.ExpectQuery("select id, category, message, status, created_at, updated_at from support").
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "category", "message", "status", "created_at", "updated_at"}))
		supports, err := repository.GetAllByUser(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, emptySupports, supports)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSupportPostgres_UpdateStatus(t *testing.T) {
	now := time.Now()
	support := models.SupportModel{
		Id:        0,
		UserId:    1,
		Category:  "test",
		Message:   "some text",
		Status:    "OPEN",
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectExec("update support").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err := repository.UpdateStatus(context.Background(), support.Id, "CLOSED")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		repository, mock := newSupportPostgres(t)
		mock.ExpectExec("update support").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(pgx.ErrNoRows)
		err := repository.UpdateStatus(context.Background(), support.Id, "...")
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
