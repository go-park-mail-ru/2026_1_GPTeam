package repository

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTransactionRepository_Create_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	row := NewMockRow(ctrl)
	repo := NewTransactionPostgres(db)

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := models.TransactionModel{
		UserId:          7,
		AccountId:       55,
		Value:           3850,
		Type:            "expense",
		Category:        "food",
		Title:           "Покупка продуктов",
		Description:     "Перекрёсток",
		TransactionDate: txDate,
	}

	db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), tx.UserId, tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate).Return(row)
	row.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = 42
		return nil
	})

	id, err := repo.Create(context.Background(), tx)
	require.NoError(t, err)
	require.Equal(t, 42, id)
}

func TestTransactionRepository_Create_UniqueViolation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	row := NewMockRow(ctrl)
	repo := NewTransactionPostgres(db)

	tx := models.TransactionModel{
		UserId:          7,
		AccountId:       55,
		Value:           1,
		Type:            "expense",
		Category:        "food",
		Title:           "Обед",
		Description:     "Кафе",
		TransactionDate: time.Now(),
	}

	db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
	row.EXPECT().Scan(gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

	id, err := repo.Create(context.Background(), tx)
	require.ErrorIs(t, err, TransactionDuplicatedDataError)
	require.Equal(t, -1, id)
}

func TestTransactionRepository_GetIdsByUserId_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	rows := NewMockRows(ctrl)
	repo := NewTransactionPostgres(db)

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

	ids, err := repo.GetIdsByUserId(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, []int{10, 11}, ids)
}

func TestTransactionRepository_Update_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	repo := NewTransactionPostgres(db)

	txDate := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	tx := models.TransactionModel{
		Id:              42,
		UserId:          7,
		AccountId:       55,
		Value:           100,
		Type:            "expense",
		Category:        "food",
		Title:           "Обед",
		Description:     "Кафе",
		TransactionDate: txDate,
	}

	db.EXPECT().Exec(gomock.Any(), gomock.Any(), tx.AccountId, tx.Value, tx.Type, tx.Category, tx.Title, tx.Description, tx.TransactionDate, tx.Id, tx.UserId).Return(pgconn.NewCommandTag("UPDATE 1"), nil)

	err := repo.Update(context.Background(), tx)
	require.NoError(t, err)
}

func TestTransactionRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	row := NewMockRow(ctrl)
	repo := NewTransactionPostgres(db)

	db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), 42).Return(row)
	row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)

	id, err := repo.Delete(context.Background(), 42)
	require.ErrorIs(t, err, NothingInTableError)
	require.Zero(t, id)
}

func TestTransactionRepository_Detail_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	db := NewMockTransactionDB(ctrl)
	row := NewMockRow(ctrl)
	repo := NewTransactionPostgres(db)

	createdAt := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	txDate := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)

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

	tx, err := repo.Detail(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, 42, tx.Id)
	require.Equal(t, 7, tx.UserId)
	require.Equal(t, 55, tx.AccountId)
	require.Equal(t, 3850.0, tx.Value)
	require.Equal(t, "Покупка продуктов", tx.Title)
	require.Equal(t, updatedAt, tx.UpdatedAt)
}
