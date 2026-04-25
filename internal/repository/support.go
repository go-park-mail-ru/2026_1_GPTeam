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
	"go.uber.org/zap"
)

type SupportRepository interface {
	Create(ctx context.Context, model models.SupportModel) (int, error)
	GetById(ctx context.Context, id int) (models.SupportModel, error)
	GetAll(ctx context.Context) ([]models.SupportModel, error)
	GetAllByUser(ctx context.Context, userId int) ([]models.SupportModel, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}

type SupportPostgres struct {
	db DB
}

func NewPostgresSupport(db DB) *SupportPostgres {
	return &SupportPostgres{
		db: db,
	}
}

func (obj *SupportPostgres) Create(ctx context.Context, support models.SupportModel) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `insert into support(user_id, category, message) values ($1, $2, $3) returning id;`
	args := []any{
		support.UserId,
		support.Category,
		support.Message,
	}
	var id int
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create support (db error)",
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
		log.Error("failed to create support (not db error)",
			zap.Error(err))
		return -1, err
	}
	log.Info("Query executed")
	return id, nil
}

func (obj *SupportPostgres) GetById(ctx context.Context, id int) (models.SupportModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select user_id, category, message, status, created_at, updated_at from support where id = $1 and deleted = false;`
	args := []any{id}
	support := models.SupportModel{
		Id:      id,
		Deleted: false,
	}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&support.UserId, &support.Category, &support.Message, &support.Status, &support.CreatedAt, &support.UpdatedAt)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get support (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.SupportModel{}, NothingInTableError
		}
		return models.SupportModel{}, err
	}
	log.Info("Query executed")
	return support, nil
}

func (obj *SupportPostgres) GetAll(ctx context.Context) ([]models.SupportModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, user_id, category, message, status, created_at, updated_at from support where deleted = false;`
	args := []any{}
	var supports []models.SupportModel
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get supports (not db error)",
			zap.Error(err))
		return []models.SupportModel{}, err
	}
	defer rows.Close()
	for rows.Next() {
		supportItem := models.SupportModel{
			Deleted: false,
		}
		err = rows.Scan(&supportItem.Id, &supportItem.UserId, &supportItem.CreatedAt, &supportItem.Message, &supportItem.Status, &supportItem.CreatedAt, &supportItem.UpdatedAt)
		if err != nil {
			log.Error("failed to scan support while getting all supports",
				zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []models.SupportModel{}, InvalidDataInTableError
			}
			return supports, err
		}
		supports = append(supports, supportItem)
	}
	if err = rows.Err(); err != nil {
		log.Error("failed to get supports",
			zap.Error(err))
		return []models.SupportModel{}, err
	}
	log.Info("Query executed")
	return supports, nil
}

func (obj *SupportPostgres) GetAllByUser(ctx context.Context, userId int) ([]models.SupportModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, category, message, status, created_at, updated_at from support where deleted = false and user_id = $1;`
	args := []any{userId}
	var supports []models.SupportModel
	startTime := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to get supports by user (not db error)",
			zap.Int("user_id", userId),
			zap.Error(err))
		return []models.SupportModel{}, err
	}
	defer rows.Close()
	for rows.Next() {
		supportItem := models.SupportModel{
			UserId:  userId,
			Deleted: false,
		}
		err = rows.Scan(&supportItem.Id, &supportItem.CreatedAt, &supportItem.Message, &supportItem.Status, &supportItem.CreatedAt, &supportItem.UpdatedAt)
		if err != nil {
			log.Error("failed to scan support while getting supports by user",
				zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return []models.SupportModel{}, InvalidDataInTableError
			}
			return supports, err
		}
		supports = append(supports, supportItem)
	}
	if err = rows.Err(); err != nil {
		log.Error("failed to get supports by user",
			zap.Error(err))
		return []models.SupportModel{}, err
	}
	log.Info("Query executed")
	return supports, nil
}

func (obj *SupportPostgres) UpdateStatus(ctx context.Context, id int, status string) error {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `update support set status = $1 where id = $2 and deleted = false;`
	args := []any{status, id}
	startTime := time.Now()
	_, err := obj.db.Exec(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	if err != nil {
		log.Error("failed to update status of support (not db error)",
			zap.Error(err))
		return err
	}
	log.Info("Query executed")
	return nil
}
