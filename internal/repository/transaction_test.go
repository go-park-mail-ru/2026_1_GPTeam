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

func TestTransactionRepository_Create(t *testing.T) {
	t.Parallel()

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	genericErr := errors.New("db down")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockTransactionDB, row *repomocks.MockRow)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), tx.UserId, tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate).Return(row)
				row.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*(dest[0].(*int)) = 42
					return nil
				})
			},
			wantID: 42,
		},
		{
			name: "unique violation",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			wantID:  -1,
			wantErr: TransactionDuplicatedDataError,
		},
		{
			name: "check violation",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			wantID:  -1,
			wantErr: ConstraintError,
		},
		{
			name: "foreign key violation",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			wantID:  -1,
			wantErr: TransactionAccountForeignKeyError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(genericErr)
			},
			wantID:  -1,
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockTransactionDB(ctrl)
			row := repomocks.NewMockRow(ctrl)
			repo := NewTransactionPostgres(db)
			tt.setupFunc(db, row)

			id, err := repo.Create(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.wantID, id)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
		})
	}
}

func TestTransactionRepository_GetIdsByUserId(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("rows broken")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows)
		wantIDs   []int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows) {
				db.EXPECT().Query(gomock.Any(), gomock.Any(), 7).Return(rows, nil)
				rows.EXPECT().Close()
				gomock.InOrder(
					rows.EXPECT().Next().Return(true),
					rows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int)) = 10
						return nil
					}),
					rows.EXPECT().Next().Return(true),
					rows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int)) = 11
						return nil
					}),
					rows.EXPECT().Next().Return(false),
				)
				rows.EXPECT().Err().Return(nil)
			},
			wantIDs: []int{10, 11},
		},
		{
			name: "query error",
			setupFunc: func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows) {
				db.EXPECT().Query(gomock.Any(), gomock.Any(), 7).Return(nil, genericErr)
			},
			wantErr: genericErr,
		},
		{
			name: "invalid data in table",
			setupFunc: func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows) {
				db.EXPECT().Query(gomock.Any(), gomock.Any(), 7).Return(rows, nil)
				rows.EXPECT().Close()
				gomock.InOrder(
					rows.EXPECT().Next().Return(true),
					rows.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows),
				)
			},
			wantErr: InvalidDataInTableError,
		},
		{
			name: "rows error",
			setupFunc: func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows) {
				db.EXPECT().Query(gomock.Any(), gomock.Any(), 7).Return(rows, nil)
				rows.EXPECT().Close()
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(genericErr)
			},
			wantErr: genericErr,
		},
		{
			name: "empty result",
			setupFunc: func(db *repomocks.MockTransactionDB, rows *repomocks.MockRows) {
				db.EXPECT().Query(gomock.Any(), gomock.Any(), 7).Return(rows, nil)
				rows.EXPECT().Close()
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(nil)
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockTransactionDB(ctrl)
			rows := repomocks.NewMockRows(ctrl)
			repo := NewTransactionPostgres(db)
			tt.setupFunc(db, rows)

			ids, err := repo.GetIdsByUserId(context.Background(), 7)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantIDs, ids)
		})
	}
}

func TestTransactionRepository_Update(t *testing.T) {
	t.Parallel()

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	genericErr := errors.New("exec failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockTransactionDB)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.NewCommandTag("UPDATE 0"), nil)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "incorrect rows affected",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.NewCommandTag("UPDATE 2"), nil)
			},
			wantErr: IncorrectRowsAffectedError,
		},
		{
			name: "foreign key violation",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			wantErr: TransactionAccountForeignKeyError,
		},
		{
			name: "check violation",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			wantErr: ConstraintError,
		},
		{
			name: "duplicated data",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			wantErr: DuplicatedDataError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockTransactionDB) {
				db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.CommandTag{}, genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockTransactionDB(ctrl)
			repo := NewTransactionPostgres(db)
			tt.setupFunc(db)

			err := repo.Update(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionRepository_Delete(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("delete failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockTransactionDB, row *repomocks.MockRow)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*(dest[0].(*int)) = 42
					return nil
				})
			},
			wantID: 42,
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any()).Return(genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockTransactionDB(ctrl)
			row := repomocks.NewMockRow(ctrl)
			repo := NewTransactionPostgres(db)
			tt.setupFunc(db, row)

			id, err := repo.Delete(context.Background(), 42)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
		})
	}
}

func TestTransactionRepository_Detail(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	txDate := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	genericErr := errors.New("detail failed")

	tests := []struct {
		name      string
		setupFunc func(db *repomocks.MockTransactionDB, row *repomocks.MockRow)
		wantTx    models.TransactionModel
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(dest ...any) error {
						*(dest[0].(*int)) = 7
						*(dest[1].(*int)) = 55
						*(dest[2].(*float64)) = 3850
						*(dest[3].(*string)) = "expense"
						*(dest[4].(*string)) = "food"
						*(dest[5].(*string)) = "Покупка продуктов"
						*(dest[6].(*string)) = "Перекрёсток"
						*(dest[7].(*time.Time)) = createdAt
						*(dest[8].(*time.Time)) = txDate
						*(dest[9].(*time.Time)) = updatedAt
						return nil
					},
				)
			},
			wantTx: models.TransactionModel{
				Id:              42,
				UserId:          7,
				AccountId:       55,
				Value:           3850,
				Type:            "expense",
				Category:        "food",
				Title:           "Покупка продуктов",
				Description:     "Перекрёсток",
				CreatedAt:       createdAt,
				TransactionDate: txDate,
				UpdatedAt:       updatedAt,
			},
		},
		{
			name: "not found",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			setupFunc: func(db *repomocks.MockTransactionDB, row *repomocks.MockRow) {
				db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
				row.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := repomocks.NewMockTransactionDB(ctrl)
			row := repomocks.NewMockRow(ctrl)
			repo := NewTransactionPostgres(db)
			tt.setupFunc(db, row)

			got, err := repo.Detail(context.Background(), 42)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantTx, got)
		})
	}
}
