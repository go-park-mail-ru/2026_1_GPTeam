package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, transactionId int) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
}

type TransactionPostgres struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewTransactionPostgres(db *pgxpool.Pool) *TransactionPostgres {
	return &TransactionPostgres{
		db:  db,
		log: logger.GetLogger(),
	}
}

func (obj *TransactionPostgres) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	obj.log.Info("creating transaction in db")
	query := `insert into transaction (user_id, account_id, value, type, category, title, description, created_at, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transaction.UserId, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Title, transaction.Description, transaction.CreatedAt, transaction.TransactionDate).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		obj.log.Error("failed to create transaction (db error)", zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return -1, TransactionDuplicatedDataError
		case pgerrcode.CheckViolation:
			return -1, ConstraintError
		case pgerrcode.ForeignKeyViolation:
			return -1, TransactionAccountForeignKeyError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		obj.log.Error("failed to create transaction (not db error)", zap.Error(err))
		return -1, err
	}
	obj.log.Info("creating transaction query executed")
	return id, nil
}

func (obj *TransactionPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	obj.log.Info("getting transaction ids by user in db")
	query := `select id from transaction where user_id = $1 and deleted_at is null;`
	var ids []int
	rows, err := obj.db.Query(ctx, query, userId)
	if err != nil {
		obj.log.Error("failed to get transaction ids by user (not db error)", zap.Error(err))
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			obj.log.Error("failed to scan id while getting transaction ids by user", zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []int{}, InvalidDataInTableError
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		obj.log.Error("failed to get transaction ids by user", zap.Error(err))
		return []int{}, err
	}
	if len(ids) == 0 {
		obj.log.Warn("no transactions found by user", zap.Int("userId", userId))
		return []int{}, NothingInTableError
	}
	obj.log.Info("get transaction ids by user query executed")
	return ids, nil
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	obj.log.Info("deleting transaction in db")
	query := `UPDATE transaction SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&id)
	if err != nil {
		obj.log.Error("failed to delete transaction (not db error)", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, NothingInTableError
		}
		return 0, err
	}
	obj.log.Info("delete transaction query executed")
	return id, nil
}

func (obj *TransactionPostgres) Detail(ctx context.Context, transactionId int) (models.TransactionModel, error) {
	obj.log.Info("getting transaction in db")
	query := `select user_id, account_id, value, type, category, title, description, created_at, transaction_date from transaction where id = $1 and deleted_at is null;`
	transaction := models.TransactionModel{Id: transactionId}
	err := obj.db.QueryRow(ctx, query, transactionId).Scan(
		&transaction.UserId, &transaction.AccountId, &transaction.Value,
		&transaction.Type, &transaction.Category, &transaction.Title,
		&transaction.Description, &transaction.CreatedAt, &transaction.TransactionDate,
	)
	if err != nil {
		obj.log.Error("failed to get transaction (not db error)", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	obj.log.Info("get transaction query executed")
	return transaction, nil
}
