package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type AccountRepository interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error)
	GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error)
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)
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
	log := logger.GetLoggerWIthRequestId(ctx)
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
	log := logger.GetLoggerWIthRequestId(ctx)
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
	log := logger.GetLoggerWIthRequestId(ctx)
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
	log := logger.GetLoggerWIthRequestId(ctx)
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
