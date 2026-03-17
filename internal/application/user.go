package application

import (
	"context"
	"fmt"
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
}

type User struct {
	repository repository.UserRepository
}

func NewUser(repository repository.UserRepository) *User {
	return &User{repository: repository}
}

func (obj *User) Create(ctx context.Context, userRequest web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error) {
	avatarUrl := "img/123.png"                                                                  // ToDo: set default in db
	bytes, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost) // ToDo: add pepper
	if err != nil {
		fmt.Printf("unable to hash password: %v\n", err)
		return web_helpers.AuthUser{}, HashPasswordError
	}
	userModel := models.UserModel{
		Id:        0,
		Username:  userRequest.Username,
		Password:  string(bytes),
		Email:     userRequest.Email,
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
		AvatarUrl: avatarUrl,
		UpdatedAt: time.Now(),
	}
	id, err := obj.repository.Create(ctx, userModel)
	if err != nil {
		return web_helpers.AuthUser{}, err
	}
	user := web_helpers.AuthUser{
		ID:        id,
		Username:  userModel.Username,
		Email:     userModel.Email,
		LastLogin: userModel.LastLogin,
		CreatedAt: userModel.CreatedAt,
	}
	return user, nil
}

func (obj *User) GetById(ctx context.Context, id int) (models.UserModel, error) {
	user, err := obj.repository.GetById(ctx, id)
	if err != nil {
		return models.UserModel{}, err
	}
	return user, nil
}

func (obj *User) GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models.UserModel, error) {
	storedUser, err := obj.repository.GetByUsername(ctx, user.Username)
	if err != nil {
		return models.UserModel{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
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
		fmt.Printf("Error while getting user by id: %v\n", err)
		return web_helpers.User{}, false
	}
	user := web_helpers.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       storedUser.LastLogin,
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	return user, true
}
