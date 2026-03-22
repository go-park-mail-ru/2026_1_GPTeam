package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface {
	Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error)
	GetById(ctx context.Context, id int) (models.UserModel, error)
	GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models.UserModel, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool)
	UpdateLastLogin(ctx context.Context, userId int) error
	Update(ctx context.Context, userInfo models.UserModel) (models.UserModel, error)
}

type User struct {
	repository repository.UserRepository
}

func NewUser(repository repository.UserRepository) *User {
	return &User{repository: repository}
}

func strPtr(s string) *string {
	return &s
}

func (obj *User) Create(ctx context.Context, userRequest web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return web_helpers.AuthUser{}, HashPasswordError
	}
	hashedPassword := string(bytes)
	userModel := models.UserModel{
		Id:        0,
		Username:  strPtr(userRequest.Username),
		Password:  strPtr(hashedPassword),
		Email:     strPtr(userRequest.Email),
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
		AvatarUrl: strPtr(""),
		UpdatedAt: time.Now(),
		Active:    true,
	}
	id, err := obj.repository.Create(ctx, userModel)
	if err != nil {
		return web_helpers.AuthUser{}, err
	}
	user := web_helpers.AuthUser{
		Id:        id,
		Username:  *userModel.Username,
		Email:     *userModel.Email,
		LastLogin: userModel.LastLogin,
		CreatedAt: userModel.CreatedAt,
	}
	return user, nil
}

func (obj *User) GetById(ctx context.Context, id int) (models.UserModel, error) {
	return obj.repository.GetById(ctx, id)
}

func (obj *User) GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models.UserModel, error) {
	storedUser, err := obj.repository.GetByUsername(ctx, user.Username)
	if err != nil {
		return models.UserModel{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(*storedUser.Password), []byte(user.Password))
	if err != nil {
		return models.UserModel{}, err
	}
	return storedUser, nil
}

func (obj *User) IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool) {
	if !isAuth {
		return web_helpers.User{}, false
	}
	storedUser, err := obj.repository.GetById(ctx, userId)
	if err != nil {
		return web_helpers.User{}, false
	}
	user := web_helpers.User{
		Username:  *storedUser.Username,
		Email:     *storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: *storedUser.AvatarUrl,
	}
	return user, true
}

func (obj *User) UpdateLastLogin(ctx context.Context, userId int) error {
	err := obj.repository.UpdateLastLogin(ctx, userId, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (obj *User) Update(ctx context.Context, userInfo models.UserModel) (models.UserModel, error) {
	if userInfo.Password != nil {
		bytes, err := bcrypt.GenerateFromPassword([]byte(*userInfo.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.UserModel{}, HashPasswordError
		}
		hashedPassword := string(bytes)
		userInfo.Password = &hashedPassword
	}
	return obj.repository.Update(ctx, userInfo)
}
