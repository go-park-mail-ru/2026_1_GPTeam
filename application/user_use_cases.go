package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

type UserUseCaseInterface interface {
	Create(ctx context.Context, user base.SignupBodyRequest) (base.AuthUser, error)
	GetById(ctx context.Context, id int) (storage.UserInfo, error)
	GetByCredentials(ctx context.Context, user base.LoginBodyRequest) (storage.UserInfo, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userID string) (base.User, bool)
}

type UserUseCase struct {
	repo repository.UserRepositoryInterface
}

func NewUserUseCases(repo repository.UserRepositoryInterface) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (obj *UserUseCase) Create(ctx context.Context, user base.SignupBodyRequest) (base.AuthUser, error) {
	avatarUrl := "img/123.png" // ToDo: set default
	newUser := storage.UserInfo{
		Id:              0,
		Username:        user.Username,
		Password:        user.Password,
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

func (obj *UserUseCase) GetById(ctx context.Context, id int) (storage.UserInfo, error) {
	user, err := obj.repo.GetById(ctx, id)
	if err != nil {
		return storage.UserInfo{}, err
	}
	return user, nil
}

func (obj *UserUseCase) GetByCredentials(ctx context.Context, user base.LoginBodyRequest) (storage.UserInfo, error) {
	storedUser, err := obj.repo.GetByCredentials(ctx, user.Username, user.Password)
	if err != nil {
		return storage.UserInfo{}, err
	}
	return storedUser, nil
}

func (obj *UserUseCase) IsAuthUserExists(ctx context.Context, isAuth bool, userID string) (base.User, bool) {
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
