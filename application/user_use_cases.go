package application

import "github.com/go-park-mail-ru/2026_1_GPTeam/repository"

type UserUseCaseInterface interface {
	Create()
	Get()
}

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Create() {

}

func (uc *UserUseCase) Get() {

}
