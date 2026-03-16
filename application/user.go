package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/web/web_helpers"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	repo repository.UserRepositoryInterface
}

func NewUser(repo repository.UserRepositoryInterface) *User {
	return &User{repo: repo}
}

func (obj *User) Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error) {
	avatarUrl := "img/123.png" // ToDo: set default
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("unable to hash password: %v\n", err)
		return web_helpers.AuthUser{}, err
	}
	newUser := models.UserInfo{
		Id:        0,
		Username:  user.Username,
		Password:  string(bytes),
		Email:     user.Email,
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
		AvatarUrl: avatarUrl,
		UpdatedAt: time.Now(),
	}
	id, err := obj.repo.Create(ctx, newUser)
	if err != nil {
		return web_helpers.AuthUser{}, err
	}
	resultUser := web_helpers.AuthUser{
		ID:        id,
		Username:  newUser.Username,
		Email:     newUser.Email,
		LastLogin: newUser.LastLogin,
		CreatedAt: newUser.CreatedAt,
	}
	return resultUser, nil
}

func (obj *User) GetById(ctx context.Context, id int) (models.UserInfo, error) {
	user, err := obj.repo.GetById(ctx, id)
	if err != nil {
		return models.UserInfo{}, err
	}
	return user, nil
}

func (obj *User) GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models.UserInfo, error) {
	storedUser, err := obj.repo.GetByUsername(ctx, user.Username)
	if err != nil {
		return models.UserInfo{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		return models.UserInfo{}, err
	}
	return storedUser, nil
}

func (obj *User) IsAuthUserExists(ctx context.Context, isAuth bool, userID int) (web_helpers.User, bool) {
	if !isAuth {
		return web_helpers.User{}, false
	}
	storedUser, err := obj.repo.GetById(ctx, userID)
	if err != nil {
		fmt.Printf("Error while getting user by id: %v\n", err)
		return web_helpers.User{}, false
	}
	authUser := web_helpers.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       storedUser.LastLogin,
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	return authUser, true
}
