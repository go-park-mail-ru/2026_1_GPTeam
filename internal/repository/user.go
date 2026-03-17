package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, userInfo models.UserModel) (int, error)
	GetById(ctx context.Context, id int) (models.UserModel, error)
	GetByUsername(ctx context.Context, username string) (models.UserModel, error)
	GetByEmail(ctx context.Context, email string) (models.UserModel, error)
}

type UserPostgres struct {
	db *pgx.Conn
}

func NewPostgresUser(db *pgx.Conn) *UserPostgres {
	return &UserPostgres{db: db}
}

func (obj *UserPostgres) Create(ctx context.Context, userInfo models.UserModel) (int, error) {
	query := `insert into "user" (username, password, email, last_login, avatar_url) VALUES ($1, $2, $3, $4, $5) returning id;`
	var id int
	username := pgtype.Text{
		String: userInfo.Username,
		Valid:  true,
	}
	password := pgtype.Text{
		String: userInfo.Password,
		Valid:  true,
	}
	email := pgtype.Text{
		String: userInfo.Email,
		Valid:  true,
	}
	lastLogin := pgtype.Timestamp{
		Time:  time.Time{},
		Valid: false,
	}
	avatarUrl := pgtype.Text{
		String: userInfo.AvatarUrl,
		Valid:  true,
	}
	err := obj.db.QueryRow(ctx, query, username, password, email, lastLogin, avatarUrl).Scan(&id)
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
		fmt.Printf("Unable to create user: %v\n", err)
		return -1, err
	}
	return id, nil
}

func (obj *UserPostgres) GetById(ctx context.Context, id int) (models.UserModel, error) {
	query := `select username, password, email, created_at, last_login, avatar_url, updated_at from "user" where id = $1;`
	var username pgtype.Text
	var password pgtype.Text
	var email pgtype.Text
	var createdAt pgtype.Timestamp
	var lastLogin pgtype.Timestamp
	var avatarUrl pgtype.Text
	var updatedAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, id).Scan(&username, &password, &email, &createdAt, &lastLogin, &avatarUrl, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user := models.UserModel{
		Id:        id,
		Username:  username.String,
		Password:  password.String,
		Email:     email.String,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		UpdatedAt: updatedAt.Time,
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserPostgres) GetByUsername(ctx context.Context, username string) (models.UserModel, error) {
	query := `select id, password, email, created_at, last_login, avatar_url, updated_at from "user" where username = $1;`
	var id pgtype.Int4
	var password pgtype.Text
	var email pgtype.Text
	var createdAt pgtype.Timestamp
	var lastLogin pgtype.Timestamp
	var avatarUrl pgtype.Text
	var updatedAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, username).Scan(&id, &password, &email, &createdAt, &lastLogin, &avatarUrl, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user := models.UserModel{
		Id:        int(id.Int32),
		Username:  username,
		Password:  password.String,
		Email:     email.String,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		UpdatedAt: updatedAt.Time,
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserPostgres) GetByEmail(ctx context.Context, email string) (models.UserModel, error) {
	query := `select id, username, password, created_at, last_login, avatar_url, updated_at from "user" where email = $1;`
	var id pgtype.Int4
	var username pgtype.Text
	var password pgtype.Text
	var createdAt pgtype.Timestamp
	var lastLogin pgtype.Timestamp
	var avatarUrl pgtype.Text
	var updatedAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, email).Scan(&id, &username, &password, &createdAt, &lastLogin, &avatarUrl, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserModel{}, NothingInTableError
		}
		fmt.Printf("Unable to get user: %v\n", err)
		return models.UserModel{}, err
	}
	user := models.UserModel{
		Id:        int(id.Int32),
		Username:  username.String,
		Password:  password.String,
		Email:     email,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		UpdatedAt: updatedAt.Time,
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}
