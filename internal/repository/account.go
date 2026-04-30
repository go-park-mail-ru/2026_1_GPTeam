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
	GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error)
	GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)
	GetById(ctx context.Context, id int) (models.AccountModel, error)
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

func (obj *AccountPostgres) Create(ctx context.Context, account models.AccountModel) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `insert into account (name, balance, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) returning id;`
	args := []any{account.Name, account.Balance, account.Currency, account.CreatedAt, account.UpdatedAt}
	var id int
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create account (db error)",
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return -1, AccountDuplicatedDataError
		case pgerrcode.CheckViolation:
			return -1, ConstraintError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		log.Error("failed to create account (not db error)",
			zap.Error(err))
		return -1, err
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
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to link account and user (db error)",
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.CheckViolation:
			return -1, ConstraintError
		case pgerrcode.ForeignKeyViolation:
			return -1, AccountForeignKeyError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		log.Error("failed to link account and user (not db error)",
			zap.Error(err))
		return -1, err
	}
	log.Info("Query executed")
	return id, nil
}

func (obj *AccountPostgres) GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id from account_user where user_id = $1 and account_id = $2`
	args := []any{userId, accountId}
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get account ids by user & account in db",
			zap.Error(err))
		return []int{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			log.Error("failed to scan id while getting account ids by user & account in db",
				zap.Error(err))
			return []int{}, UnableToGetAccountUserIdsError
		}
		ids = append(ids, id)
	}
	log.Info("Query executed")
	return ids, nil
}

func (obj *AccountPostgres) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `SELECT account_id FROM account_user WHERE user_id = $1 LIMIT 1`
	args := []any{userId}
	var accountId int
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&accountId)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get account_id by user",
			zap.Error(err))
		return 0, err
	}
	return accountId, nil
}

func (obj *AccountPostgres) GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select account.id, name, balance, currency, account.created_at, account.updated_at, coalesce(income, 0) as income, coalesce(expenses, 0) as expenses
from account join account_user on account.id = account_user.account_id left join (
select account_id, sum(case when transaction.type = 'INCOME' then transaction.value else 0 end) as income, sum(case when transaction.type = 'EXPENSE' then transaction.value else 0 end) as expenses
from transaction where deleted_at is null and transaction_date >= date_trunc('month', now()) group by account_id
) transactions on account.id = transactions.account_id
where account_user.user_id = $1;`
	args := []any{userId}
	var accounts []models.AccountModel
	var incomes, expenses []float64
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get accounts by user",
			zap.Error(err))
		return []models.AccountModel{}, []float64{}, []float64{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	for rows.Next() {
		var account models.AccountModel
		var income, expense float64
		if err = rows.Scan(&account.Id, &account.Name, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt, &income, &expense); err != nil {
			log.Error("failed to get accounts by user",
				zap.Error(err))
			return []models.AccountModel{}, []float64{}, []float64{}, UnableToGetAccountUserIdsError
		}
		accounts = append(accounts, account)
		incomes = append(incomes, income)
		expenses = append(expenses, expense)
	}
	if len(accounts) == 0 {
		log.Warn("no accounts found",
			zap.Error(NothingInTableError))
		return []models.AccountModel{}, []float64{}, []float64{}, NothingInTableError
	}
	log.Info("Query executed")
	return accounts, incomes, expenses, nil
}

func (obj *AccountPostgres) GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select account.id, name, balance, currency, created_at, updated_at from account join account_user on account.id = account_user.account_id where user_id = $1;`
	args := []any{userId}
	var accounts []models.AccountModel
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get accounts by user",
			zap.Error(err))
		return []models.AccountModel{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	for rows.Next() {
		var account models.AccountModel
		err = rows.Scan(&account.Id, &account.Name, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt)
		if err != nil {
			log.Error("failed to get accounts by user",
				zap.Error(err))
			return []models.AccountModel{}, UnableToGetAccountUserIdsError
		}
		accounts = append(accounts, account)
	}
	if len(accounts) == 0 {
		log.Warn("no accounts found",
			zap.Error(NothingInTableError))
		return []models.AccountModel{}, NothingInTableError
	}
	log.Info("Query executed")
	return accounts, nil
}

func (obj *AccountPostgres) GetById(ctx context.Context, id int) (models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select name, balance, currency, created_at, updated_at from account where id = $1;`
	args := []any{id}
	account := models.AccountModel{
		Id: id,
	}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&account.Name, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get account (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountModel{}, NothingInTableError
		}
		return models.AccountModel{}, err
	}
	log.Info("Query executed")
	return account, nil
}

func (obj *AccountPostgres) GetCurrencyByAccountId(ctx context.Context, accountId int) (string, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select currency from account where id = $1;`
	args := []any{accountId}
	var currency string
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&currency)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get currency by account",
			zap.Error(err))
		return "", err
	}
	log.Info("Query executed")
	return currency, nil
}
