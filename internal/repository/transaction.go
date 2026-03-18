package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type TransactionRepository interface {
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
}

type TransactionPostgres struct {
	db *pgx.Conn
}

func NewTransactionPostgres(db *pgx.Conn) *TransactionPostgres {
	return &TransactionPostgres{db: db}
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
