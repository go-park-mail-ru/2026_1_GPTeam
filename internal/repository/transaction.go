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
	log := logger.GetLoggerWithRequestId(ctx)
	var totalDuration time.Duration
	dbTransaction, err := obj.db.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return -1, err
	}
	defer func() {
		err = dbTransaction.Rollback(ctx)
		if err != nil {
			log.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()
	query := `insert into transaction (user_id, account_id, value, type, category, title, description, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id;`
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
	var id int
	startTime := time.Now()
	err = dbTransaction.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(startTime)
	totalDuration += duration
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
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
	log.Info("Query executed")
	log = logger.GetLoggerWithRequestId(ctx)
	query = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
	args = []any{
		transaction.Type,
		transaction.Value,
		transaction.AccountId,
	}
	startTime = time.Now()
	_, err = dbTransaction.Exec(ctx, query, args...)
	duration = time.Since(startTime)
	totalDuration += duration
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to update account (db error)",
			zap.Int("account_id", transaction.AccountId),
			zap.Int("user_id", transaction.UserId),
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return -1, TransactionAccountForeignKeyError
		case pgerrcode.CheckViolation:
			return -1, ConstraintError
		case pgerrcode.UniqueViolation:
			return -1, DuplicatedDataError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		log.Error("failed to update account (not db error)",
			zap.Int("account_id", transaction.AccountId),
			zap.Int("user_id", transaction.UserId),
			zap.Error(err))
		return -1, err
	}
	log.Info("Query executed")
	log = logger.GetLoggerWithRequestId(ctx)
	err = dbTransaction.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return -1, err
	}
	log.Info("Transaction commited", zap.String("duration", duration.String()))
	return id, nil
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
	query := `select value, type, account_id from transaction where id = $1;`
	args := []any{transaction.Id}
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
		if oldAccountId != transaction.AccountId {
			log = logger.GetLoggerWithRequestId(ctx)
			query = `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) where id = $3;`
			args = []any{
				oldType,
				oldValue,
				oldAccountId,
			}
			startTime = time.Now()
			_, err = dbTransaction.Exec(ctx, query, args...)
			duration = time.Since(startTime)
			totalDuration += duration
			log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
			pgErr, ok = errors.AsType[*pgconn.PgError](err)
			if ok {
				log.Error("failed to update account (db error)",
					zap.Int("account_id", transaction.AccountId),
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
				log.Error("failed to update account (not db error)",
					zap.Int("account_id", transaction.AccountId),
					zap.Int("user_id", transaction.UserId),
					zap.Error(err))
				return err
			}
			log.Info("Query executed")
			log = logger.GetLoggerWithRequestId(ctx)
			query = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
			args = []any{
				transaction.Type,
				transaction.Value,
				transaction.AccountId,
			}
			startTime = time.Now()
			_, err = dbTransaction.Exec(ctx, query, args...)
			duration = time.Since(startTime)
			totalDuration += duration
			log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
			pgErr, ok = errors.AsType[*pgconn.PgError](err)
			if ok {
				log.Error("failed to update account (db error)",
					zap.Int("account_id", transaction.AccountId),
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
				log.Error("failed to update account (not db error)",
					zap.Int("account_id", transaction.AccountId),
					zap.Int("user_id", transaction.UserId),
					zap.Error(err))
				return err
			}
			log.Info("Query executed")
		} else {
			if oldType != transaction.Type {
				log = logger.GetLoggerWithRequestId(ctx)
				query = `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) + (case when $3 = 'INCOME' then $4 else -1 * $4 end) where id = $5;`
				args = []any{
					oldType,
					oldValue,
					transaction.Type,
					transaction.Value,
					transaction.AccountId,
				}
				startTime = time.Now()
				_, err = dbTransaction.Exec(ctx, query, args...)
				duration = time.Since(startTime)
				totalDuration += duration
				log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
				pgErr, ok = errors.AsType[*pgconn.PgError](err)
				if ok {
					log.Error("failed to update account (db error)",
						zap.Int("account_id", transaction.AccountId),
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
					log.Error("failed to update account (not db error)",
						zap.Int("account_id", transaction.AccountId),
						zap.Int("user_id", transaction.UserId),
						zap.Error(err))
					return err
				}
				log.Info("Query executed")
			} else {
				diff := transaction.Value - oldValue
				query = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
				args = []any{
					transaction.Type,
					diff,
					transaction.AccountId,
				}
				startTime = time.Now()
				_, err = dbTransaction.Exec(ctx, query, args...)
				duration = time.Since(startTime)
				totalDuration += duration
				log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
				pgErr, ok = errors.AsType[*pgconn.PgError](err)
				if ok {
					log.Error("failed to update account (db error)",
						zap.Int("account_id", transaction.AccountId),
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
					log.Error("failed to update account (not db error)",
						zap.Int("account_id", transaction.AccountId),
						zap.Int("user_id", transaction.UserId),
						zap.Error(err))
					return err
				}
				log.Info("Query executed")
			}
		}
	}
	log = logger.GetLoggerWithRequestId(ctx)
	err = dbTransaction.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return err
	}
	log.Info("Transaction commited", zap.String("duration", duration.String()))
	return nil
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `UPDATE transaction SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING id;`
	args := []any{transactionId}
	var id int
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to delete transaction (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, NothingInTableError
		}
		return 0, err
	}
	log.Info("Query executed")
	return id, nil
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
