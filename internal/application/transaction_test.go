package application

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTransactionUseCase_Create(t *testing.T) {
	t.Parallel()

	tx := models.TransactionModel{UserId: 7, AccountId: 55, Title: "Обед"}

	tests := []struct {
		name      string
		setupFunc func(repo *repomocks.MockTransactionRepository)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Create(gomock.Any(), tx).Return(101, nil)
			},
			wantID: 101,
		},
		{
			name: "repository error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Create(gomock.Any(), tx).Return(0, repository.DuplicatedDataError)
			},
			wantErr: repository.DuplicatedDataError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockTransactionRepository(ctrl)
			uc := NewTransaction(repo)
			tt.setupFunc(repo)

			id, err := uc.Create(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, id)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
		})
	}
}

func TestTransactionUseCase_GetTransactionIdsOfUser(t *testing.T) {
	t.Parallel()

	user := models.UserModel{Id: 7}

	tests := []struct {
		name      string
		setupFunc func(repo *repomocks.MockTransactionRepository)
		wantIDs   []int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().GetIdsByUserId(gomock.Any(), 7).Return([]int{1, 2, 3}, nil)
			},
			wantIDs: []int{1, 2, 3},
		},
		{
			name: "repository error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().GetIdsByUserId(gomock.Any(), 7).Return(nil, repository.NothingInTableError)
			},
			wantErr: repository.NothingInTableError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockTransactionRepository(ctrl)
			uc := NewTransaction(repo)
			tt.setupFunc(repo)

			ids, err := uc.GetTransactionIdsOfUser(context.Background(), user)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, ids)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantIDs, ids)
		})
	}
}

func TestTransactionUseCase_Update(t *testing.T) {
	t.Parallel()

	tx := models.TransactionModel{Id: 15, UserId: 7, AccountId: 55, Title: "Обед"}

	tests := []struct {
		name      string
		setupFunc func(repo *repomocks.MockTransactionRepository)
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Update(gomock.Any(), tx).Return(nil)
			},
		},
		{
			name: "repository error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Update(gomock.Any(), tx).Return(repository.ConstraintError)
			},
			wantErr: repository.ConstraintError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockTransactionRepository(ctrl)
			uc := NewTransaction(repo)
			tt.setupFunc(repo)

			err := uc.Update(context.Background(), tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionUseCase_Delete(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("boom")

	tests := []struct {
		name      string
		setupFunc func(repo *repomocks.MockTransactionRepository)
		wantID    int
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{Id: 5, UserId: 7}, nil)
				repo.EXPECT().Delete(gomock.Any(), 5).Return(5, nil)
			},
			wantID: 5,
		},
		{
			name: "detail error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{}, repository.NothingInTableError)
			},
			wantErr: repository.NothingInTableError,
		},
		{
			name: "forbidden",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{Id: 5, UserId: 8}, nil)
			},
			wantErr: ForbiddenError,
		},
		{
			name: "delete error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 5).Return(models.TransactionModel{Id: 5, UserId: 7}, nil)
				repo.EXPECT().Delete(gomock.Any(), 5).Return(0, genericErr)
			},
			wantErr: genericErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockTransactionRepository(ctrl)
			uc := NewTransaction(repo)
			tt.setupFunc(repo)

			id, err := uc.Delete(context.Background(), 5, 7)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, id)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
		})
	}
}

func TestTransactionUseCase_Detail(t *testing.T) {
	t.Parallel()

	wantTx := models.TransactionModel{Id: 9, UserId: 7, Title: "Обед"}

	tests := []struct {
		name      string
		setupFunc func(repo *repomocks.MockTransactionRepository)
		wantTx    models.TransactionModel
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 9).Return(wantTx, nil)
			},
			wantTx: wantTx,
		},
		{
			name: "detail error",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 9).Return(models.TransactionModel{}, repository.NothingInTableError)
			},
			wantErr: repository.NothingInTableError,
		},
		{
			name: "forbidden",
			setupFunc: func(repo *repomocks.MockTransactionRepository) {
				repo.EXPECT().Detail(gomock.Any(), 9).Return(models.TransactionModel{Id: 9, UserId: 99}, nil)
			},
			wantErr: ForbiddenError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockTransactionRepository(ctrl)
			uc := NewTransaction(repo)
			tt.setupFunc(repo)

			got, err := uc.Detail(context.Background(), 9, 7)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, models.TransactionModel{}, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantTx, got)
		})
	}
}
