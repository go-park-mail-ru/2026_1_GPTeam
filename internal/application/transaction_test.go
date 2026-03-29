package application

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTransactionUseCase_Delete_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockTransactionRepository(ctrl)
	uc := NewTransaction(repo)

	repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{Id: 5, UserId: 7}, nil)
	repo.EXPECT().Delete(gomock.Any(), 5).Return(5, nil)

	id, err := uc.Delete(context.Background(), 5, 7)
	require.NoError(t, err)
	require.Equal(t, 5, id)
}

func TestTransactionUseCase_Delete_Forbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockTransactionRepository(ctrl)
	uc := NewTransaction(repo)

	repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{Id: 5, UserId: 8}, nil)

	id, err := uc.Delete(context.Background(), 5, 7)
	require.ErrorIs(t, err, ForbiddenError)
	require.Zero(t, id)
}

func TestTransactionUseCase_Detail_Forbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockTransactionRepository(ctrl)
	uc := NewTransaction(repo)

	repo.EXPECT().Detail(gomock.Any(), 9).Return(models.TransactionModel{Id: 9, UserId: 99}, nil)

	_, err := uc.Detail(context.Background(), 9, 7)
	require.ErrorIs(t, err, ForbiddenError)
}

func TestTransactionUseCase_GetTransactionIdsOfUser_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockTransactionRepository(ctrl)
	uc := NewTransaction(repo)

	repo.EXPECT().GetIdsByUserId(gomock.Any(), 7).Return([]int{1, 2, 3}, nil)

	ids, err := uc.GetTransactionIdsOfUser(context.Background(), models.UserModel{Id: 7})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, ids)
}

func TestTransactionUseCase_Update_DelegatesToRepository(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockTransactionRepository(ctrl)
	uc := NewTransaction(repo)

	tx := models.TransactionModel{Id: 15, UserId: 7, AccountId: 55, Title: "Обед"}
	repo.EXPECT().Update(gomock.Any(), tx).Return(nil)

	err := uc.Update(context.Background(), tx)
	require.NoError(t, err)
}
