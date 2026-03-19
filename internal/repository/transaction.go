package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
}

type TransactionPostgres struct {
	db *pgx.Conn
}

func NewTransactionPostgres(db *pgx.Conn) *TransactionPostgres {
	return &TransactionPostgres{db: db}
}

func (obj *TransactionPostgres) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	query := `insert into transaction (user_id, account_id, value, type, category, title, description, created_at, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`
	var id int
	userId := pgtype.Int4{
		Int32: int32(transaction.UserId),
		Valid: true,
	}
	accountId := pgtype.Int4{
		Int32: int32(transaction.AccountId),
		Valid: true,
	}
	value := pgtype.Int4{
		Int32: int32(transaction.Value),
		Valid: true,
	}
	typeArg := pgtype.Text{
		String: transaction.Type,
		Valid:  true,
	}
	category := pgtype.Text{
		String: transaction.Category,
		Valid:  true,
	}
	title := pgtype.Text{
		String: transaction.Title,
		Valid:  true,
	}
	description := pgtype.Text{
		String: transaction.Description,
		Valid:  true,
	}
	createdAt := pgtype.Timestamp{
		Time:  transaction.CreatedAt,
		Valid: true,
	}
	transactionDate := pgtype.Timestamp{
		Time:  transaction.TransactionDate,
		Valid: true,
	}
	err := obj.db.QueryRow(ctx, query, userId, accountId, value, typeArg, category, title, description, createdAt, transactionDate).Scan(&id)
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
		fmt.Printf("Unable to create transaction: %v\n", err)
		return -1, err
	}
	return id, nil
}

func (obj *TransactionPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	query := `select id from transaction where user_id = $1;`
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
