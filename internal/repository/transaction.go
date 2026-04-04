package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Update(ctx context.Context, transaction models.TransactionModel) error
	Delete(ctx context.Context, transactionId int) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
}

type TransactionPostgres struct {
	db DB
}

func NewTransactionPostgres(db DB) *TransactionPostgres {
	return &TransactionPostgres{
		db: db,
	}
}

func (obj *TransactionPostgres) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `insert into transaction (user_id, account_id, value, type, category, currency, title, description, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transaction.UserId, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Currency, transaction.Title, transaction.Description, transaction.TransactionDate).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create transaction (db error)",
			zap.Error(pgErr))
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
		log.Error("failed to create transaction (not db error)",
			zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (obj *TransactionPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id from transaction where user_id = $1 and deleted_at is null;`
	var ids []int
	rows, err := obj.db.Query(ctx, query, userId)
	if err != nil {
		log.Error("failed to get transaction ids by user (not db error)",
			zap.Error(err))
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Error("failed to scan id while getting transaction ids by user",
				zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []int{}, InvalidDataInTableError
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		log.Error("failed to get transaction ids by user",
			zap.Error(err))
		return []int{}, err
	}
	if len(ids) == 0 {
		log.Warn("no transactions found by user",
			zap.Int("userId", userId))
		return []int{}, NothingInTableError
	}
	return ids, nil
}

func (obj *TransactionPostgres) Update(ctx context.Context, transaction models.TransactionModel) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `update transaction set (account_id, value, type, category, currency, title, description, transaction_date) = ($1, $2, $3, $4, $5, $6, $7, $8) where id = $9 and user_id = $10 and deleted_at is null;`
	res, err := obj.db.Exec(ctx, query, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Currency, transaction.Title, transaction.Description, transaction.TransactionDate, transaction.Id, transaction.UserId)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to update transaction (db error)",
			zap.Int("transaction_id", transaction.Id),
			zap.Int("user_id", transaction.UserId),
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return TransactionAccountForeignKeyError
		case pgerrcode.CheckViolation:
			return ConstraintError
		case pgerrcode.UniqueViolation:
			return DuplicatedDataError
		default:
			return pgErr
		}
	}
	if err != nil {
		log.Error("failed to update transaction (not db error)",
			zap.Int("transaction_id", transaction.Id),
			zap.Int("user_id", transaction.UserId),
			zap.Error(err))
		return err
	}
	if res.RowsAffected() == 0 {
		log.Warn("failed to update transaction (no rows affected)",
			zap.Int("transaction_id", transaction.Id),
			zap.Int("user_id", transaction.UserId))
		return NothingInTableError
	}
	if res.RowsAffected() != 1 {
		log.Warn("failed to update transaction (too many rows affected)",
			zap.Int("transaction_id", transaction.Id),
			zap.Int("user_id", transaction.UserId))
		return IncorrectRowsAffectedError
	}
	return nil
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `UPDATE transaction SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&id)
	if err != nil {
		log.Error("failed to delete transaction (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, NothingInTableError
		}
		return 0, err
	}
	return id, nil
}

func (obj *TransactionPostgres) Detail(ctx context.Context, transactionId int) (models.TransactionModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select user_id, account_id, value, type, category, currency, title, description, created_at, transaction_date, updated_at from transaction where id = $1 and deleted_at is null;`
	transaction := models.TransactionModel{
		Id: transactionId,
	}

	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&transaction.UserId, &transaction.AccountId, &transaction.Value, &transaction.Type, &transaction.Category, &transaction.Currency, &transaction.Title, &transaction.Description, &transaction.CreatedAt, &transaction.TransactionDate, &transaction.UpdatedAt)
	if err != nil {
		log.Error("failed to get transaction (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	return transaction, nil
}
