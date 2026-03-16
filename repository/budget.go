package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type BudgetRepositoryInterface interface {
	Create(ctx context.Context, budget models.BudgetInfo) (int, error)
	GetById(ctx context.Context, id int) (models.BudgetInfo, error)
	GetIDsByUserId(ctx context.Context, userID int) ([]int, error)
	Delete(ctx context.Context, id int) error
}

type PostgresBudget struct {
	db *pgx.Conn
}

func NewPostgresBudget(db *pgx.Conn) *PostgresBudget {
	return &PostgresBudget{
		db: db,
	}
}

func (obj *PostgresBudget) Create(ctx context.Context, budget models.BudgetInfo) (int, error) {
	query := `insert into budget (title, description, end_at, actual, target, currency, author) VALUES ($1, $2, $3, $4, $5, $6, $7) returning id;`
	var id int
	title := pgtype.Text{
		String: budget.Title,
		Valid:  true,
	}
	description := pgtype.Text{
		String: budget.Description,
		Valid:  true,
	}
	endAt := pgtype.Timestamptz{
		Time:  budget.EndAt,
		Valid: !budget.EndAt.IsZero(),
	}
	actual := pgtype.Float8{
		Float64: float64(budget.Actual), // ToDo: change to float
		Valid:   true,
	}
	target := pgtype.Float8{
		Float64: float64(budget.Target),
		Valid:   true,
	}
	//currency := pgtype.EnumCodec{} // ToDo: to enum type (when load db)
	currency := pgtype.Text{
		String: budget.Currency,
		Valid:  true,
	}
	author := pgtype.Int4{
		Int32: int32(budget.Author),
		Valid: true,
	}
	err := obj.db.QueryRow(ctx, query, title, description, endAt, actual, target, currency, author).Scan(&id)
	if err != nil {
		fmt.Printf("Unable to create budget: %v\n", err)
		return -1, err
	}
	return id, nil
}

func (obj *PostgresBudget) GetById(ctx context.Context, id int) (models.BudgetInfo, error) {
	query := `select title, description, created_at, start_at, end_at, updated_at, actual, target, currency, author from budget where id = $1;`
	var title pgtype.Text
	var description pgtype.Text
	var createdAt pgtype.Timestamptz
	var startAt pgtype.Timestamptz
	var endAt pgtype.Timestamptz
	var updatedAt pgtype.Timestamptz
	var actual pgtype.Float8
	var target pgtype.Float8
	var currency pgtype.Text
	var author pgtype.Int4
	err := obj.db.QueryRow(ctx, query, id).Scan(&title, &description, &createdAt, &startAt, &endAt, &updatedAt, &actual, &target, &currency, &author)
	if err != nil {
		fmt.Printf("Unable to get budget: %v\n", err)
		return models.BudgetInfo{}, err
	}
	budget := models.BudgetInfo{
		Id:          id,
		Title:       title.String,
		Description: description.String,
		CreatedAt:   createdAt.Time,
		StartAt:     startAt.Time,
		EndAt:       endAt.Time,
		Actual:      actual.Float64,
		Target:      target.Float64,
		Currency:    currency.String,
		Author:      int(author.Int32),
	}
	if !endAt.Valid {
		budget.EndAt = time.Time{}
	}
	return budget, nil
}

func (obj *PostgresBudget) GetIDsByUserId(ctx context.Context, userID int) ([]int, error) {
	query := `select id from budget where author = $1;`
	var ids []int
	rows, err := obj.db.Query(ctx, query, userID)
	if err != nil {
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return []int{}, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return []int{}, err
	}
	return ids, nil
}

func (obj *PostgresBudget) Delete(ctx context.Context, id int) error {
	// ToDo: weak delete
	query := `delete from budget where id = $1;`
	_, err := obj.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
