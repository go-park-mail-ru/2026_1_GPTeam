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

func TestAccountUseCase_CreateForUser(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(11, nil)
		repo.EXPECT().LinkAccountAndUser(gomock.Any(), 11, 7).Return(0, nil)
		repo.EXPECT().GetById(gomock.Any(), 7, 11).
			Return(models.AccountModel{Id: 11, Name: "Wallet"}, nil)

		uc := NewAccount(repo)
		got, err := uc.CreateForUser(context.Background(), 7, models.AccountCreateModel{Name: "Wallet"})

		require.NoError(t, err)
		require.Equal(t, 11, got.Id)
	})

	t.Run("create fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(0, errors.New("db"))

		uc := NewAccount(repo)
		got, err := uc.CreateForUser(context.Background(), 7, models.AccountCreateModel{})

		require.Error(t, err)
		require.Equal(t, models.AccountModel{}, got)
	})

	t.Run("link fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(11, nil)
		repo.EXPECT().LinkAccountAndUser(gomock.Any(), 11, 7).Return(0, errors.New("db"))

		uc := NewAccount(repo)
		got, err := uc.CreateForUser(context.Background(), 7, models.AccountCreateModel{})

		require.Error(t, err)
		require.Equal(t, models.AccountModel{}, got)
	})
}

func TestAccountUseCase_GetById(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().GetById(gomock.Any(), 1, 2).
			Return(models.AccountModel{Id: 2, Name: "X"}, nil)

		uc := NewAccount(repo)
		got, err := uc.GetById(context.Background(), 1, 2)

		require.NoError(t, err)
		require.Equal(t, "X", got.Name)
	})

	t.Run("not found -> ErrAccountNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().GetById(gomock.Any(), 1, 2).
			Return(models.AccountModel{}, repository.NothingInTableError)

		uc := NewAccount(repo)
		_, err := uc.GetById(context.Background(), 1, 2)

		require.ErrorIs(t, err, ErrAccountNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().GetById(gomock.Any(), 1, 2).
			Return(models.AccountModel{}, errors.New("boom"))

		uc := NewAccount(repo)
		_, err := uc.GetById(context.Background(), 1, 2)

		require.Error(t, err)
	})
}

func TestAccountUseCase_GetByUserId(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockAccountRepository(ctrl)
	repo.EXPECT().GetByUserId(gomock.Any(), 5).
		Return([]models.AccountModel{{Id: 1}, {Id: 2}}, nil)

	uc := NewAccount(repo)
	got, err := uc.GetByUserId(context.Background(), 5)

	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestAccountUseCase_Update(t *testing.T) {
	t.Parallel()
	name := "new"

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), 1, 2, gomock.Any()).
			Return(models.AccountModel{Id: 2, Name: name}, nil)

		uc := NewAccount(repo)
		got, err := uc.Update(context.Background(), 1, 2, models.AccountUpdateModel{Name: &name})

		require.NoError(t, err)
		require.Equal(t, name, got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), 1, 2, gomock.Any()).
			Return(models.AccountModel{}, repository.NothingInTableError)

		uc := NewAccount(repo)
		_, err := uc.Update(context.Background(), 1, 2, models.AccountUpdateModel{Name: &name})

		require.ErrorIs(t, err, ErrAccountNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), 1, 2, gomock.Any()).
			Return(models.AccountModel{}, errors.New("boom"))

		uc := NewAccount(repo)
		_, err := uc.Update(context.Background(), 1, 2, models.AccountUpdateModel{Name: &name})

		require.Error(t, err)
	})
}

func TestAccountUseCase_Delete(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Delete(gomock.Any(), 1, 2).Return(nil)

		uc := NewAccount(repo)
		require.NoError(t, uc.Delete(context.Background(), 1, 2))
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Delete(gomock.Any(), 1, 2).Return(repository.NothingInTableError)

		uc := NewAccount(repo)
		require.ErrorIs(t, uc.Delete(context.Background(), 1, 2), ErrAccountNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		repo := repomocks.NewMockAccountRepository(ctrl)
		repo.EXPECT().Delete(gomock.Any(), 1, 2).Return(errors.New("boom"))

		uc := NewAccount(repo)
		require.Error(t, uc.Delete(context.Background(), 1, 2))
	})
}

func TestAccountUseCase_GetCurrencyByAccountId(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockAccountRepository(ctrl)
	repo.EXPECT().GetCurrencyByAccountId(gomock.Any(), 9).Return("RUB", nil)

	uc := NewAccount(repo)
	got, err := uc.GetCurrencyByAccountId(context.Background(), 9)

	require.NoError(t, err)
	require.Equal(t, "RUB", got)
}
