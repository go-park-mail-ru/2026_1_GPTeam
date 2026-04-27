package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

//go:generate mockgen -source=transaction.go -destination=mocks/transaction.go -package=mocks
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
	var id int
	err := pgx.BeginFunc(ctx, obj.db, func(dbTransaction pgx.Tx) error {
		log := logger.GetLoggerWithRequestId(ctx)
		var totalDuration time.Duration
		query := `insert into transaction (user_id, account_id, value, type, category, title, description, transaction_date) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id;`
		args := []any{
			transaction.UserId,
			transaction.AccountId,
			transaction.Value,
			transaction.Type,
			transaction.Category,
			transaction.Title,
			transaction.Description,
			transaction.TransactionDate,
		}
		startTime := time.Now()
		err := dbTransaction.QueryRow(ctx, query, args...).Scan(&id)
		duration := time.Since(startTime)
		totalDuration += duration
		log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			log.Error("failed to create transaction (db error)",
				zap.Error(pgErr))
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return TransactionDuplicatedDataError
			case pgerrcode.CheckViolation:
				return ConstraintError
			case pgerrcode.ForeignKeyViolation:
				return TransactionAccountForeignKeyError
			default:
				return pgErr
			}
		}
		if err != nil {
			log.Error("failed to create transaction (not db error)",
				zap.Error(err))
			return err
		}
		log.Info("Query executed")
		query = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
		args = []any{
			transaction.Type,
			transaction.Value,
			transaction.AccountId,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		if err != nil {
			return err
		}
		query = `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		args = []any{
			transaction.Type,
			transaction.Value,
			transaction.UserId,
			transaction.Category,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		if err != nil {
			return err
		}
		log.Info("Transaction committed", zap.String("duration", duration.String()))
		return nil
	})
	return id, err
}

func (obj *TransactionPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id from transaction where user_id = $1 and deleted_at is null;`
	args := []any{userId}
	var ids []int
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
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
	log.Info("Query executed")
	return ids, nil
}

func (obj *TransactionPostgres) Update(ctx context.Context, transaction models.TransactionModel) error {
	log := logger.GetLoggerWithRequestId(ctx)
	var totalDuration time.Duration
	dbTransaction, err := obj.db.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		err = dbTransaction.Rollback(ctx)
		if err != nil {
			log.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()
	query := `select value, type, account_id from transaction where id = $1 and deleted_at is null and user_id = $2;`
	args := []any{transaction.Id, transaction.UserId}
	var oldValue float64
	var oldType string
	var oldAccountId int
	startTime := time.Now()
	err = dbTransaction.QueryRow(ctx, query, args...).Scan(&oldValue, &oldType, &oldAccountId)
	duration := time.Since(startTime)
	totalDuration += duration
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get old transaction (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return NothingInTableError
		}
		return err
	}
	log.Info("Query executed")
	query = `update transaction set (account_id, value, type, category, title, description, transaction_date) = ($1, $2, $3, $4, $5, $6, $7) where id = $8 and user_id = $9 and deleted_at is null;`
	args = []any{
		transaction.AccountId,
		transaction.Value,
		transaction.Type,
		transaction.Category,
		transaction.Title,
		transaction.Description,
		transaction.TransactionDate,
		transaction.Id,
		transaction.UserId,
	}
	startTime = time.Now()
	res, err := dbTransaction.Exec(ctx, query, args...)
	duration = time.Since(startTime)
	totalDuration += duration
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
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
	log.Info("Query executed")
	if oldValue != transaction.Value || oldType != transaction.Type || oldAccountId != transaction.AccountId {
		duration, err = updateBalance(ctx, dbTransaction, oldValue, oldType, oldAccountId, transaction)
		if err != nil {
			return err
		}
	}
	log = logger.GetLoggerWithRequestId(ctx)
	err = dbTransaction.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return err
	}
	log.Info("Transaction committed", zap.String("duration", duration.String()))
	return nil
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, obj.db, func(dbTransaction pgx.Tx) error {
		log := logger.GetLoggerWithRequestId(ctx)
		var totalDuration time.Duration
		query := `UPDATE transaction SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING id, type, value, account_id, user_id, category;`
		args := []any{transactionId}
		var transactionType string
		var transactionValue float64
		var accountId int
		var userId int
		var category string
		startTime := time.Now()
		err := dbTransaction.QueryRow(ctx, query, args...).Scan(&id, &transactionType, &transactionValue, &accountId, &userId, &category)
		duration := time.Since(startTime)
		log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
		if err != nil {
			log.Error("failed to delete transaction (not db error)",
				zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return NothingInTableError
			}
			return err
		}
		log.Info("Query executed")
		log = logger.GetLoggerWithRequestId(ctx)
		query = `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) where id = $3;`
		args = []any{
			transactionType,
			transactionValue,
			accountId,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		if err != nil {
			return err
		}
		query = `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then -1 * $2 else $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		args = []any{
			transactionType,
			transactionValue,
			userId,
			category,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		if err != nil {
			return err
		}
		log.Info("Transaction committed", zap.String("duration", duration.String()))
		return nil
	})
	return id, err
}

func (obj *TransactionPostgres) Detail(ctx context.Context, transactionId int) (models.TransactionModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select user_id, account_id, value, type, category, title, description, created_at, transaction_date, updated_at from transaction where id = $1 and deleted_at is null;`
	args := []any{transactionId}
	transaction := models.TransactionModel{
		Id: transactionId,
	}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&transaction.UserId, &transaction.AccountId, &transaction.Value, &transaction.Type, &transaction.Category, &transaction.Title, &transaction.Description, &transaction.CreatedAt, &transaction.TransactionDate, &transaction.UpdatedAt)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get transaction (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	log.Info("Query executed")
	return transaction, nil
}

func execBalanceChangeQuery(ctx context.Context, dbTransaction pgx.Tx, query string, args ...any) (time.Duration, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	startTime := time.Now()
	_, err := dbTransaction.Exec(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to update balance (db error)",
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return duration, TransactionAccountForeignKeyError
		case pgerrcode.CheckViolation:
			return duration, ConstraintError
		case pgerrcode.UniqueViolation:
			return duration, DuplicatedDataError
		default:
			return duration, pgErr
		}
	}
	if err != nil {
		log.Error("failed to update balance (not db error)",
			zap.Error(err))
		return duration, err
	}
	log.Info("Query executed")
	return duration, nil
}

func updateBalance(ctx context.Context, dbTransaction pgx.Tx, oldValue float64, oldType string, oldAccountId int, transaction models.TransactionModel) (time.Duration, error) {
	var totalDuration time.Duration
	if oldAccountId != transaction.AccountId {
		query := `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) where id = $3;`
		args := []any{
			oldType,
			oldValue,
			oldAccountId,
		}
		duration, err := execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		if err != nil {
			return totalDuration, err
		}
		query = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
		args = []any{
			transaction.Type,
			transaction.Value,
			transaction.AccountId,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		return totalDuration, err
	}
	if oldType != transaction.Type {
		query := `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) + (case when $3 = 'INCOME' then $4 else -1 * $4 end) where id = $5;`
		args := []any{
			oldType,
			oldValue,
			transaction.Type,
			transaction.Value,
			transaction.AccountId,
		}
		duration, err := execBalanceChangeQuery(ctx, dbTransaction, query, args...)
		totalDuration += duration
		return totalDuration, err
	}
	diff := transaction.Value - oldValue
	query := `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
	args := []any{
		transaction.Type,
		diff,
		transaction.AccountId,
	}
	duration, err := execBalanceChangeQuery(ctx, dbTransaction, query, args...)
	totalDuration += duration
	return totalDuration, err
}
