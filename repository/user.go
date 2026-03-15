package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, userInfo storage.UserInfo) (int, error)
	GetById(ctx context.Context, id int) (storage.UserInfo, error)
	GetByUsername(ctx context.Context, username string) (storage.UserInfo, error)
	GetByEmail(ctx context.Context, email string) (storage.UserInfo, error)
	GetByCredentials(ctx context.Context, username string, password string) (storage.UserInfo, error)
}

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *UserRepository {
	return &UserRepository{db: db}
}

func (obj *UserRepository) Create(ctx context.Context, userInfo storage.UserInfo) (int, error) {
	query := `insert into "user" (username, password, email, last_login, avatar_url) VALUES ($1, $2, $3, $4, $5) returning id;`
	var id int
	username := pgtype.Text{
		String: userInfo.Username,
		Valid:  true,
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Unable to hash password: %v\n", err)
		return -1, err
	}
	password := pgtype.Text{
		String: string(bytes),
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
	err = obj.db.QueryRow(ctx, query, username, password, email, lastLogin, avatarUrl).Scan(&id)
	if err != nil {
		fmt.Printf("Unable to create user: %v\n", err)
		return -1, err // ToDo: add err check
	}
	return id, nil
}

func (obj *UserRepository) GetById(ctx context.Context, id int) (storage.UserInfo, error) {
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
		fmt.Printf("Unable to get user: %v\n", err)
		return storage.UserInfo{}, err
	}
	user := storage.UserInfo{
		Id:        id,
		Username:  username.String,
		Password:  password.String,
		Email:     email.String,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		// ToDo: updated_at
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserRepository) GetByUsername(ctx context.Context, username string) (storage.UserInfo, error) {
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
		fmt.Printf("Unable to get user: %v\n", err)
		return storage.UserInfo{}, err
	}
	user := storage.UserInfo{
		Id:        int(id.Int32),
		Username:  username,
		Password:  password.String,
		Email:     email.String,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		// ToDo: updated_at
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserRepository) GetByEmail(ctx context.Context, email string) (storage.UserInfo, error) {
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
		fmt.Printf("Unable to get user: %v\n", err) // ToDo: add errors to global constants + add pgx.ErrNoRows
		return storage.UserInfo{}, err
	}
	user := storage.UserInfo{
		Id:        int(id.Int32),
		Username:  username.String,
		Password:  password.String,
		Email:     email,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		// ToDo: updated_at
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}

func (obj *UserRepository) GetByCredentials(ctx context.Context, username string, password string) (storage.UserInfo, error) {
	query := `select id, email, created_at, last_login, avatar_url, updated_at from "user" where username = $1 and password = $2;`
	var id pgtype.Int4
	var email pgtype.Text
	var createdAt pgtype.Timestamp
	var lastLogin pgtype.Timestamp
	var avatarUrl pgtype.Text
	var updatedAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, username, password).Scan(&id, &email, &createdAt, &lastLogin, &avatarUrl, &updatedAt)
	if err != nil {
		fmt.Printf("Unable to get user: %v\n", err)
		return storage.UserInfo{}, err
	}
	user := storage.UserInfo{
		Id:        int(id.Int32),
		Username:  username,
		Password:  password,
		Email:     email.String,
		CreatedAt: createdAt.Time,
		LastLogin: lastLogin.Time,
		AvatarUrl: avatarUrl.String,
		// ToDo: updated_at
	}
	if !lastLogin.Valid {
		user.LastLogin = time.Time{}
	}
	return user, nil
}
