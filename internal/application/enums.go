package application

import (
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type EnumsUseCase interface {
	GetCurrencyCodes() []string
	GetTransactionTypes() []string
	GetCategoryTypes() []string
}

type Enums struct {
	repository repository.EnumsRepository
	log        *zap.Logger
}

func NewEnums(repository repository.EnumsRepository) *Enums {
	return &Enums{
		repository: repository,
		log:        logger.GetLogger(),
	}
}

func (obj *Enums) GetCurrencyCodes() []string {
	obj.log.Info("getting currency codes")
	return obj.repository.GetCurrencyCodesFromDB()
}

func (obj *Enums) GetTransactionTypes() []string {
	obj.log.Info("getting transaction types")
	return obj.repository.GetTransactionTypesFromDB()
}

func (obj *Enums) GetCategoryTypes() []string {
	obj.log.Info("getting category types")
	return obj.repository.GetCategoryTypesFromDB()
}
