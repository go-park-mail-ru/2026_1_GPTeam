package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresBudget struct { // ToDo: change naming
	db         *pgx.Conn
	mu         sync.RWMutex
	currencies []string
}

func getCurrenciesFromDB(db *pgx.Conn) []string {
	query := `select enumlabel from pg_enum where enumtypid = 'currency_code'::regtype order by enumsortorder;`
	row, err := db.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("unable get currencies from db: %v\n", err) // ToDo: use new error
		return []string{}
	}
	var currencies []string
	for row.Next() {
		var code string
		err = row.Scan(&code)
		if err != nil {
			fmt.Printf("unable get currencies from db: %v\n", err)
			return []string{}
		}
		currencies = append(currencies, code)
	}
	return currencies
}

func NewPostgresBudget(db *pgx.Conn) *PostgresBudget {
	currencies := getCurrenciesFromDB(db)
	fmt.Printf("Read currencies from db: %v\n", currencies)
	return &PostgresBudget{
		db:         db,
		currencies: currencies,
	}
}

func (obj *PostgresBudget) Create(ctx context.Context, budget models.BudgetModel) (int, error) {
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
		Float64: budget.Actual,
		Valid:   true,
	}
	target := pgtype.Float8{
		Float64: budget.Target,
		Valid:   true,
	}
	currency := pgtype.Text{
		String: budget.Currency,
		Valid:  true,
	}
	author := pgtype.Int4{
		Int32: int32(budget.Author),
		Valid: true,
	}
	err := obj.db.QueryRow(ctx, query, title, description, endAt, actual, target, currency, author).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		switch pgErr.Code {
		case "23505":
			return -1, DuplicatedDataError
		case "23514":
			return -1, ConstraintError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		fmt.Printf("Unable to create budget: %v\n", err)
		return -1, err
	}
	return id, nil
}

func (obj *PostgresBudget) GetById(ctx context.Context, id int) (models.BudgetModel, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BudgetModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get budget: %v\n", err)
		return models.BudgetModel{}, err
	}
	budget := models.BudgetModel{
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

func (obj *PostgresBudget) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	query := `select id from budget where author = $1;`
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

func (obj *PostgresBudget) Delete(ctx context.Context, id int) error {
	// ToDo: weak delete
	query := `delete from budget where id = $1;`
	_, err := obj.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (obj *PostgresBudget) GetCurrencies() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.currencies
}
