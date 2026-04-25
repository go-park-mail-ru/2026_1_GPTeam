package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type SupportUseCase interface {
	Create(ctx context.Context, data web_helpers.SupportRequest, userId int) (int, error)
	GetById(ctx context.Context, id int) (models.SupportModel, error)
	GetAll(ctx context.Context) ([]models.SupportModel, error)
	GetAllByUser(ctx context.Context, userId int) ([]models.SupportModel, error)
	Update(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
}

type Support struct {
	repository repository.SupportRepository
}

func NewSupport(repository repository.SupportRepository) *Support {
	return &Support{
		repository: repository,
	}
}

func (obj *Support) Create(ctx context.Context, data web_helpers.SupportRequest, userId int) (int, error) {
	support := models.SupportModel{
		Id:        0,
		UserId:    userId,
		Category:  data.Category,
		Message:   data.Message,
		Status:    "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   false,
	}
	id, err := obj.repository.Create(ctx, support)
	return id, err
}

func (obj *Support) GetById(ctx context.Context, id int) (models.SupportModel, error) {
	support, err := obj.repository.GetById(ctx, id)
	return support, err
}

func (obj *Support) GetAll(ctx context.Context) ([]models.SupportModel, error) {
	supports, err := obj.repository.GetAll(ctx)
	return supports, err
}

func (obj *Support) GetAllByUser(ctx context.Context, userId int) ([]models.SupportModel, error) {
	supports, err := obj.repository.GetAllByUser(ctx, userId)
	return supports, err
}

func (obj *Support) Update(ctx context.Context, id int, status string) error {
	err := obj.repository.UpdateStatus(ctx, id, status)
	return err
}

func (obj *Support) Delete(ctx context.Context, id int) error {
	panic("implement me")
}
