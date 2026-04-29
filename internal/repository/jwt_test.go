package repository

import (
	"context"
	"errors"
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
	expiredAt := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)
	token := models.RefreshTokenModel{Uuid: "rt-1", UserId: 7, ExpiredAt: expiredAt}
	genericErr := errors.New("insert failed")

	tests := []struct {
		name      string
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into jwt`).
					WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "unique violation",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into jwt`).
					WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			wantErr: DuplicatedDataError,
		},
		{
			name: "check violation",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into jwt`).
					WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			wantErr: ConstraintError,
		},
		{
			name: "generic error",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into jwt`).
					WithArgs(token.Uuid, token.UserId, token.ExpiredAt).
					WillReturnError(genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newJwtPostgres(t)
			tt.setupFunc(mock)

			err := repo.Create(context.Background(), token)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestJwtPostgres_DeleteByUuid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		uuid      string
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name: "success",
			uuid: "rt-1",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`delete from jwt`).
					WithArgs("rt-1").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "not found",
			uuid: "rt-1",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`delete from jwt`).
					WithArgs("rt-1").
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newJwtPostgres(t)
			tt.setupFunc(mock)

			err := repo.DeleteByUuid(context.Background(), tt.uuid)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestJwtPostgres_DeleteByUserId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userId    int
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name:   "success",
			userId: 7,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`delete from jwt`).
					WithArgs(7).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name:   "not found",
			userId: 7,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`delete from jwt`).
					WithArgs(7).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newJwtPostgres(t)
			tt.setupFunc(mock)

			err := repo.DeleteByUserId(context.Background(), tt.userId)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestJwtPostgres_Get(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name      string
		uuid      string
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name: "success",
			uuid: "rt-1",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"user_id", "expired_at"}).AddRow(7, now)
				mock.ExpectQuery(`select user_id, expired_at from jwt`).
					WithArgs("rt-1").
					WillReturnRows(rows)
			},
		},
		{
			name: "not found",
			uuid: "rt-1",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select user_id, expired_at from jwt`).
					WithArgs("rt-1").
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newJwtPostgres(t)
			tt.setupFunc(mock)

			token, err := repo.Get(context.Background(), tt.uuid)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 7, token.UserId)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
