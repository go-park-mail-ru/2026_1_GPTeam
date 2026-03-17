package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget models.BudgetModel) (int, error)
	GetById(ctx context.Context, id int) (models.BudgetModel, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, id int) error
	GetCurrencies() []string
}

type BudgetPostgres struct {
	db         *pgx.Conn
	mu         sync.RWMutex
	currencies []string
}

func getCurrenciesFromDB(db *pgx.Conn) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'currency_code'::regtype order by enumsortorder;`
	row, err := db.Query(context.Background(), query)
	if err != nil {
		return []string{}, UnableToReadCurrenciesError
	}
	var currencies []string
	for row.Next() {
		var code string
		err = row.Scan(&code)
		if err != nil {
			return []string{}, UnableToReadCurrenciesError
		}
		currencies = append(currencies, code)
	}
	return currencies, nil
}

func NewBudgetPostgres(db *pgx.Conn) (*BudgetPostgres, error) {
	currencies, err := getCurrenciesFromDB(db)
	if err != nil {
		return &BudgetPostgres{}, err
	}
	fmt.Printf("Read currencies from db: %v\n", currencies)
	return &BudgetPostgres{
		db:         db,
		currencies: currencies,
	}, nil
}

func (obj *BudgetPostgres) Create(ctx context.Context, budget models.BudgetModel) (int, error) {
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
		case pgerrcode.UniqueViolation:
			return -1, DuplicatedDataError
		case pgerrcode.CheckViolation:
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

func (obj *BudgetPostgres) GetById(ctx context.Context, id int) (models.BudgetModel, error) {
	query := `select title, description, created_at, start_at, end_at, updated_at, actual, target, currency, author, active from budget where id = $1 and active = true;`
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
	var active pgtype.Bool
	err := obj.db.QueryRow(ctx, query, id).Scan(&title, &description, &createdAt, &startAt, &endAt, &updatedAt, &actual, &target, &currency, &author, &active)
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
		Active:      active.Bool,
	}
	if !endAt.Valid {
		budget.EndAt = time.Time{}
	}
	return budget, nil
}

func (obj *BudgetPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	query := `select id from budget where author = $1 and active = true;`
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

func (obj *BudgetPostgres) Delete(ctx context.Context, id int) error {
	query := `update budget set active = false where id = $1;`
	_, err := obj.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (obj *BudgetPostgres) GetCurrencies() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.currencies
}
