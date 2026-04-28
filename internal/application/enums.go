package application

import (
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=enums.go -destination=mocks/enums.go -package=mocks
type EnumsUseCase interface {
	GetCurrencyCodes() []string
	GetTransactionTypes() []string
	GetCategoryTypes() []string
}

type Enums struct {
	repository repository.EnumsRepository
}

func NewEnums(repository repository.EnumsRepository) *Enums {
	return &Enums{
		repository: repository,
	}
}

func (obj *Enums) GetCurrencyCodes() []string {
	return obj.repository.GetCurrencyCodesFromDB()
}

func (obj *Enums) GetTransactionTypes() []string {
	return obj.repository.GetTransactionTypesFromDB()
}

func (obj *Enums) GetCategoryTypes() []string {
	return obj.repository.GetCategoryTypesFromDB()
}
