package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget models.BudgetModel) (int, error)
	GetById(ctx context.Context, id int) (models.BudgetModel, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, id int) error
}

type BudgetPostgres struct {
	db DB
}

func NewBudgetPostgres(db DB) *BudgetPostgres {
	return &BudgetPostgres{
		db: db,
	}
}

func (obj *BudgetPostgres) Create(ctx context.Context, budget models.BudgetModel) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `insert into budget (title, description, created_at, start_at, end_at, actual, target, currency, author) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`
	endAt := pgtype.Timestamptz{
		Time:  budget.EndAt,
		Valid: !budget.EndAt.IsZero(),
	}
	args := []any{
		budget.Title,
		budget.Description,
		budget.CreatedAt,
		budget.StartAt,
		endAt,
		budget.Actual,
		budget.Target,
		budget.Currency,
		budget.Author,
	}
	var id int
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create budget (db error)",
			zap.Error(pgErr))
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
		log.Error("failed to create budget (not db error)",
			zap.Error(err))
		return -1, err
	}
	log.Info("Query executed")
	return id, nil
}

func (obj *BudgetPostgres) GetById(ctx context.Context, id int) (models.BudgetModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select title, description, created_at, start_at, end_at, updated_at, actual, target, currency, author, active from budget where id = $1 and active = true;`
	var endAt pgtype.Timestamptz
	budget := models.BudgetModel{Id: id}
	args := []any{id}
	timeStart := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&budget.Title, &budget.Description, &budget.CreatedAt, &budget.StartAt, &endAt, &budget.UpdatedAt, &budget.Actual, &budget.Target, &budget.Currency, &budget.Author, &budget.Active)
	duration := time.Since(timeStart)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get budget (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BudgetModel{}, NothingInTableError
		}
		return models.BudgetModel{}, err
	}
	budget.EndAt = endAt.Time
	if !endAt.Valid {
		budget.EndAt = time.Time{}
	}
	log.Info("Query executed")
	return budget, nil
}

func (obj *BudgetPostgres) GetIdsByUserId(ctx context.Context, userId int) ([]int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id from budget where author = $1 and active = true;`
	var ids []int
	args := []any{userId}
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get budget ids by user (not db error)",
			zap.Error(err))
		return []int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Error("failed to scan id while getting budget ids by user",
				zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []int{}, InvalidDataInTableError
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		log.Error("failed to get budget ids by user (not db error)",
			zap.Error(err))
		return []int{}, err
	}
	if len(ids) == 0 {
		log.Info("no budget ids found in db")
		return []int{}, NothingInTableError
	}
	log.Info("Query executed")
	return ids, nil
}

func (obj *BudgetPostgres) Delete(ctx context.Context, id int) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `update budget set active = false where id = $1;`
	args := []any{id}
	startTime := time.Now()
	_, err := obj.db.Exec(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to delete budget (not db error)",
			zap.Error(err))
		return err
	}
	log.Info("Query executed")
	return nil
}
