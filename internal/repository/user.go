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
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, userInfo models.UserModel) (int, error)
	GetById(ctx context.Context, id int) (models.UserModel, error)
	GetByUsername(ctx context.Context, username string) (models.UserModel, error)
	GetByEmail(ctx context.Context, email string) (models.UserModel, error)
	UpdateLastLogin(ctx context.Context, userId int, lastLogin time.Time) error
}

type UserPostgres struct {
	db *pgx.Conn
}

func NewUserPostgres(db *pgx.Conn) *UserPostgres {
	return &UserPostgres{db: db}
}

func (obj *UserPostgres) Create(ctx context.Context, userInfo models.UserModel) (int, error) {
	query := `insert into "user" (username, password, email, last_login) VALUES ($1, $2, $3, $4) returning id;`
	var id int
	lastLogin := pgtype.Timestamp{
		Time:  time.Time{},
		Valid: false,
	}
	err := obj.db.QueryRow(ctx, query, userInfo.Username, userInfo.Password, userInfo.Email, lastLogin).Scan(&id)
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
		fmt.Printf("Unable to create user: %v\n", err)
		return -1, err
	}
	return id, nil
}

func (obj *UserPostgres) GetById(ctx context.Context, id int) (models.UserModel, error) {
	query := `select username, password, email, created_at, last_login, avatar_url, updated_at, active from "user" where id = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{
		Id: id,
	}
	err := obj.db.QueryRow(ctx, query, id).Scan(&user.Username, &user.Password, &user.Email, &user.CreatedAt, &lastLogin, &user.AvatarUrl, &user.UpdatedAt, &user.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user.LastLogin = lastLogin.Time
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserPostgres) GetByUsername(ctx context.Context, username string) (models.UserModel, error) {
	query := `select id, password, email, created_at, last_login, avatar_url, updated_at, active from "user" where username = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{
		Username: username,
	}
	err := obj.db.QueryRow(ctx, query, username).Scan(&user.Id, &user.Password, &user.Email, &user.CreatedAt, &lastLogin, &user.AvatarUrl, &user.UpdatedAt, &user.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user.LastLogin = lastLogin.Time
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserPostgres) GetByEmail(ctx context.Context, email string) (models.UserModel, error) {
	query := `select id, username, password, created_at, last_login, avatar_url, updated_at, active from "user" where email = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{
		Email: email,
	}
	err := obj.db.QueryRow(ctx, query, email).Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt, &lastLogin, &user.AvatarUrl, &user.UpdatedAt, &user.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user.LastLogin = lastLogin.Time
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserPostgres) UpdateLastLogin(ctx context.Context, userId int, lastLogin time.Time) error {
	query := `UPDATE "user" SET last_login = $1 WHERE id = $2;`
	_, err := obj.db.Exec(ctx, query, lastLogin, userId)
	if err != nil {
		return err
	}
	return nil
}
