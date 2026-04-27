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

type AccountRepository interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error)
	GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error)
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)
	GetById(ctx context.Context, userId int, accountId int) (models.AccountModel, error)
	GetByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)
	Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error)
	Delete(ctx context.Context, userId int, accountId int) error
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
	log := logger.GetLoggerWIthRequestId(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountNotFound
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

func (obj *AccountPostgres) Create(ctx context.Context, account models.AccountModel) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
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
	log := logger.GetLoggerWIthRequestId(ctx)
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
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id from account_user where user_id = $1 and account_id = $2`
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
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `SELECT account_id FROM account_user WHERE user_id = $1 LIMIT 1`
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
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `
		select a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at
		from account a
		join account_user au on au.account_id = a.id
		where au.user_id = $1 and a.id = $2`
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
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `
		select a.id, a.name, a.balance, a.currency, a.created_at, a.updated_at
		from account a
		join account_user au on au.account_id = a.id
		where au.user_id = $1
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

func (obj *AccountPostgres) Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `
		update account a
		set
			name = coalesce($3, a.name),
			balance = coalesce($4, a.balance),
			currency = coalesce($5, a.currency),
			updated_at = now()
		from account_user au
		where au.account_id = a.id and au.user_id = $1 and a.id = $2
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
	log := logger.GetLoggerWIthRequestId(ctx)
	tx, err := obj.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin account delete tx", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `delete from account_user where user_id = $1 and account_id = $2`, userId, accountId)
	if err != nil {
		log.Error("failed to unlink account and user", zap.Error(err))
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	_, err = tx.Exec(ctx, `delete from account where id = $1`, accountId)
	if err != nil {
		log.Error("failed to delete account", zap.Error(err))
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit account delete tx", zap.Error(err))
		return err
	}
	log.Info("account deleted", zap.Int("account_id", accountId), zap.Int("user_id", userId))
	return nil
}
