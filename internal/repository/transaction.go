package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -source=transaction.go -destination=mocks/transaction.go -package=mocks
type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Update(ctx context.Context, transaction models.TransactionModel) error
	Delete(ctx context.Context, transactionId int) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
}

type TransactionDB interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
	Config() *pgxpool.Config
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Reset()
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Stat() *pgxpool.Stat
}

type TransactionPostgres struct {
	db TransactionDB
}

func NewTransactionPostgres(db TransactionDB) *TransactionPostgres {
	return &TransactionPostgres{db: db}
}

func (obj *TransactionPostgres) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	query := `insert into transaction (user_id, account_id, value, type, category, title, description, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transaction.UserId, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Title, transaction.Description, transaction.TransactionDate).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
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
		return -1, err
	}
	return id, nil
}

func (obj *TransactionPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	query := `select id from transaction where user_id = $1 and deleted_at is null;`
	var ids []int
	rows, err := obj.db.Query(ctx, query, userId)
	if err != nil {
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return []int{}, InvalidDataInTableError
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return []int{}, err
	}
	if len(ids) == 0 {
		return []int{}, NothingInTableError
	}
	return ids, nil
}

func (obj *TransactionPostgres) Update(ctx context.Context, transaction models.TransactionModel) error {
	query := `update transaction set (account_id, value, type, category, title, description, transaction_date) = ($1, $2, $3, $4, $5, $6, $7) where id = $8 and user_id = $9 and deleted_at is null;`
	res, err := obj.db.Exec(ctx, query, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Title, transaction.Description, transaction.TransactionDate, transaction.Id, transaction.UserId)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
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
		fmt.Printf("Unable to update transaction: %v\n", err)
		return err
	}
	if res.RowsAffected() == 0 {
		return NothingInTableError
	}
	if res.RowsAffected() != 1 {
		return IncorrectRowsAffectedError
	}
	return nil
}

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	query := `UPDATE transaction SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING id;`
	var id int
	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, NothingInTableError
		}
		return 0, err
	}
	return id, nil
}

func (obj *TransactionPostgres) Detail(ctx context.Context, transactionId int) (models.TransactionModel, error) {
	query := `select user_id, account_id, value, type, category, title, description, created_at, transaction_date, updated_at from transaction where id = $1 and deleted_at is null;`
	transaction := models.TransactionModel{
		Id: transactionId,
	}

	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&transaction.UserId, &transaction.AccountId, &transaction.Value, &transaction.Type, &transaction.Category, &transaction.Title, &transaction.Description, &transaction.CreatedAt, &transaction.TransactionDate, &transaction.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	return transaction, nil
}
