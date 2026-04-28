package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
)

func TestAccountUseCase_Create(t *testing.T) {
	t.Parallel()

	account := models.AccountModel{Name: "base", Balance: 0, Currency: "RUB", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockAccountRepository)
		expectedId  int
		expectedErr bool
	}{
		{
			name: "успешное создание",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
			},
			expectedId:  1,
			expectedErr: false,
		},
		{
			name: "ошибка репозитория",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, errors.New("db error"))
			},
			expectedId:  -1,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockAccountRepository(ctrl)
			c.setupMocks(repo)

			uc := NewAccount(repo)
			id, err := uc.Create(context.Background(), account)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedId, id)
			}
		})
	}
}

func TestAccountUseCase_LinkAccountAndUser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockAccountRepository)
		expectedErr bool
	}{
		{
			name: "успешная линковка",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().LinkAccountAndUser(gomock.Any(), 1, 1).Return(1, nil)
			},
			expectedErr: false,
		},
		{
			name: "ошибка репозитория",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().LinkAccountAndUser(gomock.Any(), 1, 1).Return(-1, errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockAccountRepository(ctrl)
			c.setupMocks(repo)

			uc := NewAccount(repo)
			err := uc.LinkAccountAndUser(context.Background(), 1, 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccountUseCase_IsUserAuthorOfAccount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockAccountRepository)
		expected   bool
	}{
		{
			name: "является автором",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().GetIdsByUserAndAccount(gomock.Any(), 1, 1).Return([]int{1}, nil)
			},
			expected: true,
		},
		{
			name: "не является автором — пустой список",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().GetIdsByUserAndAccount(gomock.Any(), 1, 1).Return([]int{}, nil)
			},
			expected: false,
		},
		{
			name: "ошибка репозитория → false",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().GetIdsByUserAndAccount(gomock.Any(), 1, 1).Return([]int{}, errors.New("db error"))
			},
			expected: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockAccountRepository(ctrl)
			c.setupMocks(repo)

			uc := NewAccount(repo)
			result := uc.IsUserAuthorOfAccount(context.Background(), 1, 1)
			require.Equal(t, c.expected, result)
		})
	}
}

func TestAccountUseCase_GetAccountIdByUserId(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockAccountRepository)
		expectedId  int
		expectedErr bool
	}{
		{
			name: "успешно",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().GetAccountIdByUserId(gomock.Any(), 1).Return(42, nil)
			},
			expectedId:  42,
			expectedErr: false,
		},
		{
			name: "ошибка",
			setupMocks: func(repo *repomocks.MockAccountRepository) {
				repo.EXPECT().GetAccountIdByUserId(gomock.Any(), 1).Return(0, errors.New("db error"))
			},
			expectedId:  0,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockAccountRepository(ctrl)
			c.setupMocks(repo)

			uc := NewAccount(repo)
			id, err := uc.GetAccountIdByUserId(context.Background(), 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedId, id)
			}
		})
	}
}
