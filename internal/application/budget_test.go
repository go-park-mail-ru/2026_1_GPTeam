package application

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks" // Используем мок репозитория!
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBudget_Create(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Юзкейс тестируется через мок репозитория
	repo := mocks.NewMockBudgetRepository(ctrl)
	useCase := NewBudget(repo)

	loc := time.FixedZone("UTC+3", 3*60*60)
	now := time.Now()
	// Ожидаем, что время в Create обнулится до 00:00:00 в зоне UTC+3
	expectedTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	budget := models.BudgetModel{
		Title:     "Trip",
		CreatedAt: now,
	}

	t.Run("success", func(t *testing.T) {
		// Проверяем, что юзкейс передает в репозиторий уже "нормализованное" время
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, b models.BudgetModel) (int, error) {
			if !b.CreatedAt.Equal(expectedTime) {
				t.Errorf("expected time %v, got %v", expectedTime, b.CreatedAt)
			}
			return 1, nil
		})

		id, err := useCase.Create(context.Background(), budget)
		require.NoError(t, err)
		require.Equal(t, 1, id)
	})
}

func TestBudget_GetById(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBudgetRepository(ctrl)
	useCase := NewBudget(repo)

	user := models.UserModel{Id: 7}
	budget := models.BudgetModel{Id: 1, Author: 7, Title: "Budget"}

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().GetById(gomock.Any(), 1).Return(budget, nil)

		res, err := useCase.GetById(context.Background(), 1, user)
		require.NoError(t, err)
		require.Equal(t, budget.Title, res.Title)
	})

	t.Run("forbidden", func(t *testing.T) {
		otherUser := models.UserModel{Id: 999}
		repo.EXPECT().GetById(gomock.Any(), 1).Return(budget, nil)

		_, err := useCase.GetById(context.Background(), 1, otherUser)
		// Проверка твоей кастомной ошибки из юзкейса
		require.ErrorIs(t, err, UserNotAuthorOfBudgetError)
	})
}

func TestBudget_Delete(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBudgetRepository(ctrl)
	useCase := NewBudget(repo)

	user := models.UserModel{Id: 7}
	budget := models.BudgetModel{Id: 1, Author: 7}

	t.Run("success delete", func(t *testing.T) {
		// В коде Delete вызывает GetById, поэтому мокаем оба вызова
		repo.EXPECT().GetById(gomock.Any(), 1).Return(budget, nil)
		repo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

		err := useCase.Delete(context.Background(), 1, user)
		require.NoError(t, err)
	})

	t.Run("fail - not author", func(t *testing.T) {
		otherUser := models.UserModel{Id: 999}
		repo.EXPECT().GetById(gomock.Any(), 1).Return(budget, nil)

		err := useCase.Delete(context.Background(), 1, otherUser)
		require.ErrorIs(t, err, UserNotAuthorOfBudgetError)
	})
}

func TestBudget_GetBudgetsOfUser(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBudgetRepository(ctrl)
	useCase := NewBudget(repo)

	user := models.UserModel{Id: 7}
	expectedIds := []int{1, 2}

	repo.EXPECT().GetIdsByUserId(gomock.Any(), user.Id).Return(expectedIds, nil)

	ids, err := useCase.GetBudgetsOfUser(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, expectedIds, ids)
}
