package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) (int, error)
	GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error)
}

type AccountPostgres struct {
	db *pgxpool.Pool
}

func NewAccountPostgres(db *pgxpool.Pool) *AccountPostgres {
	return &AccountPostgres{db: db}
}

func (obj *AccountPostgres) Create(ctx context.Context, account models.AccountModel) (int, error) {
	query := `insert into account (name, balance, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, account.Name, account.Balance, account.Currency, account.CreatedAt, account.UpdatedAt).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
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
		return -1, err
	}
	return id, nil
}

func (obj *AccountPostgres) GetIdsByUserAndAccount(ctx context.Context, userId int, accountId int) ([]int, error) {
	query := `select id from account_user where user_id = $1 and account_id = $2`
	rows, err := obj.db.Query(ctx, query, userId, accountId)
	if err != nil {
		return []int{}, UnableToGetAccountUserIdsError
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return []int{}, UnableToGetAccountUserIdsError
		}
		ids = append(ids, id)
	}
	return ids, nil
}
