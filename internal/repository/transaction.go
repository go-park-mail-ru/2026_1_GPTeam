package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/currency_converter"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=transaction.go -destination=mocks/mock_transaction.go -package=mocks
type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel, account models.AccountModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Update(ctx context.Context, transaction models.TransactionModel, oldTransaction models.TransactionModel, account models.AccountModel, oldAccount models.AccountModel) error
	Delete(ctx context.Context, transactionId int, account models.AccountModel) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
	Search(ctx context.Context, userId int, filters TransactionFilters) ([]models.TransactionModel, error)
}

type TransactionFilters struct {
	StartDate      *time.Time
	EndDate        *time.Time
	Category       *string
	AccountID      *int
	SearchQuery    *string
}

type TransactionPostgres struct {
	db DB
}

func NewTransactionPostgres(db DB) *TransactionPostgres {
	return &TransactionPostgres{
		db: db,
	}
}

func (obj *TransactionPostgres) Create(ctx context.Context, transaction models.TransactionModel, account models.AccountModel) (int, error) {
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
		accQuery := `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
		accArgs := []any{
			transaction.Type,
			transaction.Value,
			transaction.AccountId,
		}
		budgetQuery := `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		budgetArgs := []any{
			transaction.Type,
			currency_converter.ConvertToRub(transaction.Value, account.Currency),
			transaction.UserId,
			transaction.Category,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
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
		log.Error("failed to get transaction ids by user (not db error)", zap.Error(err))
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Error("failed to scan id while getting transaction ids by user", zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []int{}, InvalidDataInTableError
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		log.Error("failed to get transaction ids by user", zap.Error(err))
		return []int{}, err
	}
	if len(ids) == 0 {
		log.Warn("no transactions found by user", zap.Int("userId", userId))
		return []int{}, NothingInTableError
	}
	log.Info("Query executed")
	return ids, nil
}

func (obj *TransactionPostgres) Update(ctx context.Context, transaction models.TransactionModel, oldTransaction models.TransactionModel, account models.AccountModel, oldAccount models.AccountModel) error {
	err := pgx.BeginFunc(ctx, obj.db, func(dbTransaction pgx.Tx) error {
		log := logger.GetLoggerWithRequestId(ctx)
		var totalDuration time.Duration
		query := `update transaction set (account_id, value, type, category, title, description, transaction_date) = ($1, $2, $3, $4, $5, $6, $7) where id = $8 and user_id = $9 and deleted_at is null;`
		args := []any{
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
		startTime := time.Now()
		res, err := dbTransaction.Exec(ctx, query, args...)
		duration := time.Since(startTime)
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
			if errors.Is(err, pgx.ErrNoRows) {
				return NothingInTableError
			}
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
		if oldTransaction.Category != transaction.Category || oldTransaction.Value != transaction.Value || oldTransaction.Type != transaction.Type || oldTransaction.AccountId != transaction.AccountId {
			duration, err = updateBalance(ctx, dbTransaction, oldTransaction, transaction, oldAccount, account)
			if err != nil {
				return err
			}
		}
		log.Info("Transaction committed", zap.String("duration", duration.String()))
		return nil
	})
	return err
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int, account models.AccountModel) (int, error) {
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
		accQuery := `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) where id = $3;`
		accArgs := []any{
			transactionType,
			transactionValue,
			accountId,
		}
		budgetQuery := `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then -1 * $2 else $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		budgetArgs := []any{
			transactionType,
			currency_converter.ConvertToRub(transactionValue, account.Currency),
			userId,
			category,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
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
		log.Error("failed to get transaction (not db error)", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	log.Info("Query executed")
	return transaction, nil
}

func (obj *TransactionPostgres) Search(ctx context.Context, userId int, filters TransactionFilters) ([]models.TransactionModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)

	query := `select id, user_id, account_id, value, type, category, title, description, created_at, transaction_date, updated_at from transaction where user_id = $1 and deleted_at is null`
	args := []any{userId}
	argIndex := 2

	if filters.StartDate != nil {
		query += ` and transaction_date >= $` + fmt.Sprint(argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		query += ` and transaction_date <= $` + fmt.Sprint(argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.Category != nil {
		query += ` and category = $` + fmt.Sprint(argIndex)
		args = append(args, *filters.Category)
		argIndex++
	}

	if filters.AccountID != nil {
		query += ` and account_id = $` + fmt.Sprint(argIndex)
		args = append(args, *filters.AccountID)
		argIndex++
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		query += ` and (title ILIKE $` + fmt.Sprint(argIndex) + ` or description ILIKE $` + fmt.Sprint(argIndex) + `)`
		searchPattern := "%" + *filters.SearchQuery + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	query += ` order by transaction_date desc, created_at desc`

	var transactions []models.TransactionModel
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to search transactions (not db error)", zap.Error(err))
		return []models.TransactionModel{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionModel
		err = rows.Scan(&transaction.Id, &transaction.UserId, &transaction.AccountId, &transaction.Value, &transaction.Type, &transaction.Category, &transaction.Title, &transaction.Description, &transaction.CreatedAt, &transaction.TransactionDate, &transaction.UpdatedAt)
		if err != nil {
			log.Error("failed to scan transaction while searching", zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []models.TransactionModel{}, InvalidDataInTableError
			}
			return transactions, err
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Error("failed to search transactions", zap.Error(err))
		return []models.TransactionModel{}, err
	}

	if len(transactions) == 0 {
		log.Warn("no transactions found with filters", zap.Int("userId", userId))
		return []models.TransactionModel{}, NothingInTableError
	}
	log.Info("Query executed")
	return transactions, nil
}

func execBalanceChangeQuery(ctx context.Context, dbTransaction pgx.Tx, accQuery string, accArgs []any, budgetQuery string, budgetArgs []any) (time.Duration, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	startTime := time.Now()
	_, err := dbTransaction.Exec(ctx, accQuery, accArgs...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, accQuery, accArgs, duration)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to update account (db error)",
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
		log.Error("failed to update account (not db error)",
			zap.Error(err))
		return duration, err
	}
	log.Info("Query executed")
	log = logger.GetLoggerWithRequestId(ctx)
	startTime = time.Now()
	_, err = dbTransaction.Exec(ctx, budgetQuery, budgetArgs...)
	duration += time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, budgetQuery, budgetArgs, duration)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to update budget (db error)",
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
		log.Error("failed to update budget (not db error)",
			zap.Error(err))
		return duration, err
	}
	log.Info("Query executed")
	return duration, nil
}

func updateBalance(ctx context.Context, dbTransaction pgx.Tx, oldTransaction models.TransactionModel, transaction models.TransactionModel, oldAccount models.AccountModel, account models.AccountModel) (time.Duration, error) {
	var totalDuration time.Duration
	if oldTransaction.AccountId != transaction.AccountId {
		accQuery := `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) where id = $3;`
		accArgs := []any{
			oldTransaction.Type,
			oldTransaction.Value,
			oldTransaction.AccountId,
		}
		budgetQuery := `update budget set actual = greatest(0, least(budget.target, actual + (case when $1 = 'INCOME' then -1 * $2 else $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		budgetArgs := []any{
			oldTransaction.Type,
			currency_converter.ConvertToRub(oldTransaction.Value, oldAccount.Currency),
			oldTransaction.UserId,
			oldTransaction.Category,
		}
		duration, err := execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
		totalDuration += duration
		if err != nil {
			return totalDuration, err
		}
		accQuery = `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
		accArgs = []any{
			transaction.Type,
			transaction.Value,
			transaction.AccountId,
		}
		budgetQuery = `update budget set actual = greatest(0, least(budget.target, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		budgetArgs = []any{
			transaction.Type,
			currency_converter.ConvertToRub(transaction.Value, account.Currency),
			transaction.UserId,
			transaction.Category,
		}
		duration, err = execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
		totalDuration += duration
		return totalDuration, err
	}
	if oldTransaction.Type != transaction.Type {
		accQuery := `update account set balance = balance + (case when $1 = 'INCOME' then -1 * $2 else $2 end) + (case when $3 = 'INCOME' then $4 else -1 * $4 end) where id = $5;`
		accArgs := []any{
			oldTransaction.Type,
			currency_converter.ConvertToRub(oldTransaction.Value, account.Currency),
			transaction.Type,
			currency_converter.ConvertToRub(transaction.Value, account.Currency),
			transaction.AccountId,
		}
		budgetQuery := `update budget set actual = greatest(0, least(budget.target, actual + (case when $1 = 'INCOME' then -1 * $2 else $2 end) + (case when $3 = 'INCOME' then $4 else -1 * $4 end))) where author = $5 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $6);`
		budgetArgs := []any{
			oldTransaction.Type,
			currency_converter.ConvertToRub(oldTransaction.Value, account.Currency),
			transaction.Type,
			transaction.Value,
			transaction.UserId,
			transaction.Category,
		}
		duration, err := execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
		totalDuration += duration
		return totalDuration, err
	}
	diff := transaction.Value - oldTransaction.Value
	accQuery := `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`
	accArgs := []any{
		transaction.Type,
		diff,
		transaction.AccountId,
	}
	budgetQuery := `update budget set actual = greatest(0, least(budget.target, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
	budgetArgs := []any{
		transaction.Type,
		currency_converter.ConvertToRub(diff, account.Currency),
		transaction.UserId,
		transaction.Category,
	}
	duration, err := execBalanceChangeQuery(ctx, dbTransaction, accQuery, accArgs, budgetQuery, budgetArgs)
	totalDuration += duration
	if err != nil {
		return totalDuration, err
	}
	if oldTransaction.Category != transaction.Category {
		log := logger.GetLoggerWithRequestId(ctx)
		query := `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then -1 * $2 else $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		args := []any{
			oldTransaction.Type,
			currency_converter.ConvertToRub(oldTransaction.Value, oldAccount.Currency),
			oldTransaction.UserId,
			oldTransaction.Category,
		}
		startTime := time.Now()
		_, err = dbTransaction.Exec(ctx, query, args...)
		duration = time.Since(startTime)
		totalDuration += duration
		log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
		if err != nil {
			log.Error("failed to update budget (not db error)",
				zap.Error(err))
			return totalDuration, err
		}
		log.Info("Query executed")
		log = logger.GetLoggerWithRequestId(ctx)
		query = `update budget set actual = greatest(0, least(target, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end))) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`
		args = []any{
			transaction.Type,
			currency_converter.ConvertToRub(transaction.Value, account.Currency),
			transaction.UserId,
			transaction.Category,
		}
		startTime = time.Now()
		_, err = dbTransaction.Exec(ctx, query, args...)
		duration = time.Since(startTime)
		totalDuration += duration
		log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
		if err != nil {
			log.Error("failed to update budget (not db error)",
				zap.Error(err))
			return totalDuration, err
		}
		log.Info("Query executed")
		return totalDuration, nil
	}
	return totalDuration, err
}
