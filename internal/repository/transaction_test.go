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

func newTransactionPostgres(t *testing.T) (*TransactionPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
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

func TestTransactionPostgres_Create(t *testing.T) {
	t.Parallel()
	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	genericErr := errors.New("db down")

	tests := []struct {
		name      string
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`insert into transaction`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(42))
				mock.ExpectExec(`update account set balance`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			wantID: 42,
		},
		{
			name: "unique violation",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`insert into transaction`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				mock.ExpectRollback()
			},
			wantID:  -1,
			wantErr: TransactionDuplicatedDataError,
		},
		{
			name: "check violation",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`insert into transaction`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
				mock.ExpectRollback()
			},
			wantID:  -1,
			wantErr: ConstraintError,
		},
		{
			name: "foreign key violation",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`insert into transaction`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
				mock.ExpectRollback()
			},
			wantID:  -1,
			wantErr: TransactionAccountForeignKeyError,
		},
		{
			name: "generic error",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`insert into transaction`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(genericErr)
				mock.ExpectRollback()
			},
			wantID:  -1,
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			id, err := repo.Create(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.wantID, id)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionPostgres_GetIdsByUserId(t *testing.T) {
	t.Parallel()
	genericErr := errors.New("rows broken")

	tests := []struct {
		name      string
		userId    int
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantIDs   []int
		wantErr   error
	}{
		{
			name:   "success",
			userId: 7,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(10).AddRow(11)
				mock.ExpectQuery(`select id from transaction`).
					WithArgs(7).
					WillReturnRows(rows)
			},
			wantIDs: []int{10, 11},
		},
		{
			name:   "empty result",
			userId: 7,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id from transaction`).
					WithArgs(7).
					WillReturnRows(pgxmock.NewRows([]string{"id"}))
			},
			wantErr: NothingInTableError,
		},
		{
			name:   "query error",
			userId: 7,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id from transaction`).
					WithArgs(7).
					WillReturnError(genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			ids, err := repo.GetIdsByUserId(context.Background(), tt.userId)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantIDs, ids)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionPostgres_Update(t *testing.T) {
	t.Parallel()
	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := validTransactionModel(txDate)
	genericErr := errors.New("update failed")

	tests := []struct {
		name      string
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
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
			},
		},
		{
			name: "not found",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`select value, type, account_id from transaction`).
					WithArgs(tx.Id, tx.UserId).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`select value, type, account_id from transaction`).
					WithArgs(tx.Id, tx.UserId).
					WillReturnError(genericErr)
				mock.ExpectRollback()
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			err := repo.Update(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionPostgres_Delete(t *testing.T) {
	t.Parallel()
	genericErr := errors.New("delete failed")

	tests := []struct {
		name      string
		txId      int
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			txId: 42,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE transaction SET deleted_at`).
					WithArgs(42).
					WillReturnRows(pgxmock.NewRows([]string{"id", "type", "value", "account_id"}).AddRow(42, "expense", float64(100), 1))
				mock.ExpectExec(`update account set balance`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			wantID: 42,
		},
		{
			name: "not found",
			txId: 42,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE transaction SET deleted_at`).
					WithArgs(42).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: NothingInTableError,
		},
		{
			name: "generic error",
			txId: 42,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE transaction SET deleted_at`).
					WithArgs(42).
					WillReturnError(genericErr)
				mock.ExpectRollback()
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			id, err := repo.Delete(context.Background(), tt.txId)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionPostgres_Detail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		txId      int
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name: "success",
			txId: 42,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"user_id", "account_id", "value", "type", "category", "title", "description", "created_at", "transaction_date", "updated_at"}).
					AddRow(7, 55, float64(3850), "expense", "food", "Покупка", "Описание", time.Now(), time.Now(), time.Now())
				mock.ExpectQuery(`select user_id, account_id`).
					WithArgs(42).
					WillReturnRows(rows)
			},
		},
		{
			name: "not found",
			txId: 42,
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select user_id, account_id`).
					WithArgs(42).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			tx, err := repo.Detail(context.Background(), tt.txId)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 7, tx.UserId)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionPostgres_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userId    int
		filters   TransactionFilters
		setupFunc func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name:   "success",
			userId: 7,
			filters: TransactionFilters{},
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "account_id", "value", "type", "category", "title", "description", "created_at", "transaction_date", "updated_at"}).
					AddRow(42, 7, 55, float64(3850), "expense", "food", "Покупка", "Описание", time.Now(), time.Now(), time.Now())
				mock.ExpectQuery(`select id, user_id`).
					WithArgs(7).
					WillReturnRows(rows)
			},
		},
		{
			name:   "empty result",
			userId: 7,
			filters: TransactionFilters{},
			setupFunc: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, user_id`).
					WithArgs(7).
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "account_id", "value", "type", "category", "title", "description", "created_at", "transaction_date", "updated_at"}))
			},
			wantErr: NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTransactionPostgres(t)
			tt.setupFunc(mock)

			txs, err := repo.Search(context.Background(), tt.userId, tt.filters)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, txs, 1)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
