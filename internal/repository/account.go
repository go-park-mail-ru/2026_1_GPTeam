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

//go:generate go run go.uber.org/mock/mockgen@latest -source=account.go -destination=mocks/mock_account.go -package=mocks
type AccountRepository interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error)
	GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error)
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)

	GetById(ctx context.Context, userId int, accountId int) (models.AccountModel, error)
	GetByAccountId(ctx context.Context, accountId int) (models.AccountModel, error)
	GetByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)
	GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error)
	GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)

	Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error)
	Delete(ctx context.Context, userId int, accountId int) error
	GetCurrencyByAccountId(ctx context.Context, accountId int) (string, error)
}

type AccountPostgres struct {
	db DB
}

func NewAccountPostgres(db DB) *AccountPostgres {
	return &AccountPostgres{
		db: db,
	}
}

func mapAccountPgError(ctx context.Context, err error, action string) error {
	if err == nil {
		return nil
	}
	log := logger.GetLoggerWithRequestId(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return NothingInTableError
	}
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error(action, zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return AccountDuplicatedDataError
		case pgerrcode.CheckViolation:
			return ConstraintError
		case pgerrcode.ForeignKeyViolation:
			return AccountForeignKeyError
		default:
			return pgErr
		}
	}
	log.Error(action, zap.Error(err))
	return err
}

func (obj *AccountPostgres) GetByAccountId(ctx context.Context, accountId int) (models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)

	query := `
		select id, name, balance, currency, created_at, updated_at
		from account
		where id = $1
		  and deleted_at is null`

	args := []any{accountId}

	var account models.AccountModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&account.Id,
		&account.Name,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))

	if mappedErr := mapAccountPgError(ctx, err, "failed to get account by account id"); mappedErr != nil {
		return models.AccountModel{}, mappedErr
	}

	log.Info("Query executed")
	return account, nil
}

func (obj *AccountPostgres) GetCurrencyByAccountId(ctx context.Context, accountId int) (string, error) {
	log := logger.GetLoggerWithRequestId(ctx)

	query := `
		select currency
		from account
		where id = $1
		  and deleted_at is null`

	args := []any{accountId}

	var currency string
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&currency)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))

	if mappedErr := mapAccountPgError(ctx, err, "failed to get currency by account id"); mappedErr != nil {
		return "", mappedErr
	}

	log.Info("Query executed")
	return currency, nil
}

func (obj *AccountPostgres) Create(ctx context.Context, account models.AccountModel) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `insert into account (name, balance, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) returning id;`
	args := []any{account.Name, account.Balance, account.Currency, account.CreatedAt, account.UpdatedAt}
	var id int
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if mappedErr := mapAccountPgError(ctx, err, "failed to create account"); mappedErr != nil {
		return -1, mappedErr
	}
	log.Info("Query executed")
	return id, nil
}

func (obj *AccountPostgres) LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `insert into account_user (account_id, user_id) VALUES ($1, $2) returning id;`
	args := []any{accountId, userId}
	var id int
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if mappedErr := mapAccountPgError(ctx, err, "failed to link account and user"); mappedErr != nil {
		return -1, mappedErr
	}
	log.Info("Query executed")
	return id, nil
}

func (obj *AccountPostgres) GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
	select id
	from account_user au
	where user_id = $1
	  and account_id = $2
	  and exists (
		  select 1
		  from account a
		  where a.id = au.account_id
		    and a.deleted_at is null
	  )`
	args := []any{userId, accountId}
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get account ids by user & account in db", zap.Error(err))
		return []int{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			log.Error("failed to scan id while getting account ids by user & account in db", zap.Error(err))
			return []int{}, UnableToGetAccountUserIdsError
		}
		ids = append(ids, id)
	}
	log.Info("Query executed")
	return ids, nil
}

func (obj *AccountPostgres) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
	SELECT account_id
	FROM account_user au
	WHERE user_id = $1
	  AND EXISTS (
		  SELECT 1
		  FROM account a
		  WHERE a.id = au.account_id
		    AND a.deleted_at IS NULL
	  )
	LIMIT 1`
	args := []any{userId}
	var accountId int
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&accountId)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if mappedErr := mapAccountPgError(ctx, err, "failed to get account id by user id"); mappedErr != nil {
		return -1, mappedErr
	}
	log.Info("Query executed")
	return accountId, nil
}

func (obj *AccountPostgres) GetById(ctx context.Context, userId int, accountId int) (models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		select a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at
		from account a
		join account_user au on au.account_id = a.id
		where au.user_id = $1 and a.id = $2 and a.deleted_at is null`
	args := []any{userId, accountId}
	var account models.AccountModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&account.Id,
		&account.Name,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))
	if mappedErr := mapAccountPgError(ctx, err, "failed to get account by id"); mappedErr != nil {
		return models.AccountModel{}, mappedErr
	}
	log.Info("Query executed")
	return account, nil
}

func (obj *AccountPostgres) GetByUserId(ctx context.Context, userId int) ([]models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		select a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at
		from account a
		join account_user au on au.account_id = a.id
		where au.user_id = $1 and a.deleted_at is null
		order by a.id`
	args := []any{userId}
	start := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))
	if err != nil {
		log.Error("failed to get accounts by user id", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	accounts := make([]models.AccountModel, 0)
	for rows.Next() {
		var account models.AccountModel
		if err = rows.Scan(&account.Id, &account.Name, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt); err != nil {
			log.Error("failed to scan account", zap.Error(err))
			return nil, InvalidDataInTableError
		}
		accounts = append(accounts, account)
	}
	if rows.Err() != nil {
		log.Error("failed while reading accounts", zap.Error(rows.Err()))
		return nil, rows.Err()
	}
	log.Info("Query executed")
	return accounts, nil
}

func (obj *AccountPostgres) GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error) {
	return obj.GetByUserId(ctx, userId)
}

func (obj *AccountPostgres) GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error) {
	log := logger.GetLoggerWithRequestId(ctx)

	query := `
		select
			a.id,
			a.name,
			a.balance,
			a.currency,
			a.created_at,
			a.updated_at,
			coalesce(sum(t.value) filter (where t.type = 'INCOME'), 0) as income,
			coalesce(sum(t.value) filter (where t.type = 'EXPENSE'), 0) as expense
		from account a
		join account_user au on au.account_id = a.id
		left join transaction t
			on t.account_id = a.id
			and t.user_id = au.user_id
			and t.deleted_at is null
		where au.user_id = $1
		  and a.deleted_at is null
		group by a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at
		order by a.id`

	args := []any{userId}

	start := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))

	if err != nil {
		log.Error("failed to get accounts with balance by user id", zap.Error(err))
		return nil, nil, nil, err
	}
	defer rows.Close()

	accounts := make([]models.AccountModel, 0)
	incomes := make([]float64, 0)
	expenses := make([]float64, 0)

	for rows.Next() {
		var account models.AccountModel
		var income float64
		var expense float64

		if err = rows.Scan(
			&account.Id,
			&account.Name,
			&account.Balance,
			&account.Currency,
			&account.CreatedAt,
			&account.UpdatedAt,
			&income,
			&expense,
		); err != nil {
			log.Error("failed to scan account with balance", zap.Error(err))
			return nil, nil, nil, InvalidDataInTableError
		}

		accounts = append(accounts, account)
		incomes = append(incomes, income)
		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		log.Error("failed while reading accounts with balance", zap.Error(err))
		return nil, nil, nil, err
	}

	log.Info("Query executed")

	return accounts, incomes, expenses, nil
}

func (obj *AccountPostgres) Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		update account a
		set
			name = coalesce($3, a.name),
			balance = coalesce($4, a.balance),
			currency = coalesce($5, a.currency),
			updated_at = now()
		from account_user au
		where au.account_id = a.id and au.user_id = $1 and a.id = $2 and a.deleted_at is null
		returning a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at`
	args := []any{userId, accountId, account.Name, account.Balance, account.Currency}
	var updated models.AccountModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&updated.Id,
		&updated.Name,
		&updated.Balance,
		&updated.Currency,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, time.Since(start))
	if mappedErr := mapAccountPgError(ctx, err, "failed to update account"); mappedErr != nil {
		return models.AccountModel{}, mappedErr
	}
	log.Info("Query executed")
	return updated, nil
}

func (obj *AccountPostgres) Delete(ctx context.Context, userId int, accountId int) error {
	log := logger.GetLoggerWithRequestId(ctx)

	return pgx.BeginFunc(ctx, obj.db, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		batch.Queue(`
			update transaction
			set deleted_at = now()
			where user_id = $1
			  and account_id = $2
			  and deleted_at is null`,
			userId,
			accountId,
		)

		batch.Queue(`
			update account a
			set deleted_at = now(),
			    updated_at = now()
			where a.id = $1
			  and a.deleted_at is null
			  and exists (
				  select 1
				  from account_user au
				  where au.account_id = a.id
				    and au.user_id = $2
			  )`,
			accountId,
			userId,
		)

		results := tx.SendBatch(ctx, batch)
		defer results.Close()

		if _, err := results.Exec(); err != nil {
			log.Error("failed to soft delete account transactions", zap.Error(err))
			return err
		}

		tag, err := results.Exec()
		if err != nil {
			log.Error("failed to soft delete account", zap.Error(err))
			return err
		}

		if tag.RowsAffected() == 0 {
			return NothingInTableError
		}

		log.Info(
			"account soft deleted",
			zap.Int("account_id", accountId),
			zap.Int("user_id", userId),
		)

		return nil
	})
}
