package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCaseInterface interface {
	Create(ctx context.Context, user base.SignupBodyRequest) (base.AuthUser, error)
	GetById(ctx context.Context, id int) (models.UserInfo, error)
	GetByCredentials(ctx context.Context, user base.LoginBodyRequest) (models.UserInfo, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userID string) (base.User, bool)
}

type User struct {
	repo repository.UserRepositoryInterface
}

func NewUser(repo repository.UserRepositoryInterface) *User {
	return &User{repo: repo}
}

func (obj *User) Create(ctx context.Context, user base.SignupBodyRequest) (base.AuthUser, error) {
	avatarUrl := "img/123.png" // ToDo: set default
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("unable to hash password: %v\n", err)
		return base.AuthUser{}, err
	}
	newUser := models.UserInfo{
		Id:              0,
		Username:        user.Username,
		Password:        string(bytes),
		Email:           user.Email,
		CreatedAt:       time.Now(),
		LastLogin:       time.Time{},
		AvatarUrl:       avatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB", // ToDo: delete
	}
	id, err := obj.repo.Create(ctx, newUser)
	if err != nil {
		return base.AuthUser{}, err
	}
	resultUser := base.AuthUser{
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

func (obj *User) GetByCredentials(ctx context.Context, user base.LoginBodyRequest) (models.UserInfo, error) {
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

func (obj *User) IsAuthUserExists(ctx context.Context, isAuth bool, userID string) (base.User, bool) {
	if !isAuth {
		return base.User{}, false
	}
	id, err := strconv.Atoi(userID)
	if err != nil {
		return base.User{}, false
	}
	storedUser, err := obj.repo.GetById(ctx, id)
	if err != nil {
		fmt.Printf("Error while getting user by id: %v\n", err)
		return base.User{}, false
	}
	authUser := base.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       storedUser.LastLogin,
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         storedUser.Balance,
		BalanceCurrency: storedUser.BalanceCurrency,
	}
	return authUser, true
}
