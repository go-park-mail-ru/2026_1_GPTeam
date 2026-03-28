package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type AccountRepository interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error)
	GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error)
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)
}

type AccountPostgres struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewAccountPostgres(db *pgxpool.Pool) *AccountPostgres {
	return &AccountPostgres{
		db:  db,
		log: logger.GetLogger(),
	}
}

func (obj *AccountPostgres) Create(ctx context.Context, account models.AccountModel) (int, error) {
	query := `insert into account (name, balance, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, account.Name, account.Balance, account.Currency, account.CreatedAt, account.UpdatedAt).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		obj.log.Error("failed to create account (db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
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
		obj.log.Error("failed to create account (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (obj *AccountPostgres) LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error) {
	query := `insert into account_user (account_id, user_id) VALUES ($1, $2) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, accountId, userId).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		obj.log.Error("failed to link account and user (db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
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
		obj.log.Error("failed to link account and user (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (obj *AccountPostgres) GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error) {
	query := `select id from account_user where user_id = $1 and account_id = $2`
	rows, err := obj.db.Query(ctx, query, userId, accountId)
	if err != nil {
		obj.log.Error("failed to get account ids by user & account in db",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return []int{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			obj.log.Error("failed to scan id while getting account ids by user & account in db",
				zap.String("request_id", ctx.Value("request_id").(string)),
				zap.Error(err))
			return []int{}, UnableToGetAccountUserIdsError
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (obj *AccountPostgres) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	query := `SELECT account_id FROM account_user WHERE user_id = $1 LIMIT 1`
	var accountId int
	err := obj.db.QueryRow(ctx, query, userId).Scan(&accountId)
	if err != nil {
		obj.log.Error("failed to get account_id by user",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return 0, err
	}
	return accountId, nil
}
