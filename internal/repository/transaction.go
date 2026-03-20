package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, transactionId int) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
	IsUserAuthorOfTransaction(transaction models.TransactionModel, user models.UserModel) (bool, error)
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
	err := obj.db.QueryRow(ctx, query, transaction.UserId, transaction.AccountId, transaction.Value, transaction.Type, transaction.Category, transaction.Title, transaction.Description, transaction.CreatedAt, transaction.TransactionDate).Scan(&id)
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

func (obj *TransactionPostgres) Delete(ctx context.Context, transactionId int) (int, error) {
	query := `delete from transaction where id = $1 returning id;`
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
	query := `select user_id, account_id, value, type, category, title, description, created_at, transaction_date from transaction where id = $1;`
	var userId int
	var accountId int
	var value float64
	var transactionType string
	var category string
	var title string
	var description string
	var createdAt time.Time
	var transactionDate time.Time
	err := obj.db.QueryRow(ctx, query, transactionId).Scan(&userId, &accountId, &value, &transactionType, &category, &title, &description, &createdAt, &transactionDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TransactionModel{}, NothingInTableError
		}
		return models.TransactionModel{}, err
	}
	transaction := models.TransactionModel{
		Id:              transactionId,
		UserId:          userId,
		AccountId:       accountId,
		Value:           value,
		Type:            transactionType,
		Category:        category,
		Title:           title,
		Description:     description,
		CreatedAt:       createdAt,
		TransactionDate: transactionDate,
	}
	return transaction, nil
}

func (obj *TransactionPostgres) IsUserAuthorOfTransaction(transaction models.TransactionModel, user models.UserModel) (bool, error) {
	return transaction.UserId == user.Id, nil
}
