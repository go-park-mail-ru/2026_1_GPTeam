package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

type UserUseCaseInterface interface {
	Create(ctx context.Context, user base.SignupBodyRequest) (base.AuthUser, error)
}

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
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
