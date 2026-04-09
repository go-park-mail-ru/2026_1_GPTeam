package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestJwtRepository_Create(t *testing.T) {
	t.Parallel()

	expiredAt := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)
	token := models.RefreshTokenModel{Uuid: "rt-1", UserId: 7, ExpiredAt: expiredAt}
	genericErr := errors.New("insert failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockJwtDB)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), token.Uuid, token.UserId, token.ExpiredAt).Return(pgconn.NewCommandTag("INSERT 1"), nil)
			},
		},
		{
			name: "unique violation",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), token.Uuid, token.UserId, token.ExpiredAt).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			wantErr: DuplicatedDataError,
		},
		{
			name: "check violation",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), token.Uuid, token.UserId, token.ExpiredAt).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			wantErr: ConstraintError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), token.Uuid, token.UserId, token.ExpiredAt).Return(pgconn.CommandTag{}, genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockJwtDB(ctrl)
			repo := NewJwtPostgres(db)
			tt.setupFunc(db)

			err := repo.Create(context.Background(), token)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestJwtRepository_DeleteByUuid(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("delete failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockJwtDB)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), "rt-1").Return(pgconn.NewCommandTag("DELETE 1"), nil)
			},
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), "rt-1").Return(pgconn.CommandTag{}, pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), "rt-1").Return(pgconn.CommandTag{}, genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockJwtDB(ctrl)
			repo := NewJwtPostgres(db)
			tt.setupFunc(db)

			err := repo.DeleteByUuid(context.Background(), "rt-1")
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestJwtRepository_DeleteByUserId(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("delete failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockJwtDB)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), 7).Return(pgconn.NewCommandTag("DELETE 1"), nil)
			},
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), 7).Return(pgconn.CommandTag{}, pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockJwtDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), 7).Return(pgconn.CommandTag{}, genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockJwtDB(ctrl)
			repo := NewJwtPostgres(db)
			tt.setupFunc(db)

			err := repo.DeleteByUserId(context.Background(), 7)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestJwtRepository_Get(t *testing.T) {
	t.Parallel()

	expiredAt := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)
	genericErr := errors.New("scan failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockJwtDB, row *repomocks.MockRow)
		wantToken models.RefreshTokenModel
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockJwtDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), "rt-1").Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
					*(dest[0].(*int)) = 7
					*(dest[1].(*time.Time)) = expiredAt
					return nil
				})
			},
			wantToken: models.RefreshTokenModel{Uuid: "rt-1", UserId: 7, ExpiredAt: expiredAt},
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockJwtDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), "rt-1").Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "too many rows",
			setupFunc: func(db *repomocks.MockJwtDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), "rt-1").Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(pgx.ErrTooManyRows)
			},
			wantErr: TooManyRowsError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockJwtDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), "rt-1").Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockJwtDB(ctrl)
			row := repomocks.NewMockRow(ctrl)
			repo := NewJwtPostgres(db)
			tt.setupFunc(db, row)

			got, err := repo.Get(context.Background(), "rt-1")
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantToken, got)
		})
	}
}
