package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func TestUserUseCase_GetById(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{Id: 1, Username: "testuser", Email: "test@example.com"}

	cases := []struct {
		name        string
		id          int
		setupMocks  func(repo *repomocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name: "пользователь найден",
			id:   1,
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedErr: false,
		},
		{
			name: "пользователь не найден",
			id:   999,
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), 999).Return(nil, errors.New("not found"))
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

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			user, err := uc.GetById(context.Background(), c.id)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.Equal(t, testUser.Username, user.Username)
			}
		})
	}
}

func TestUserUseCase_Create(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		req         web_helpers.SignupBodyRequest
		setupMocks  func(repo *repomocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name: "успешное создание",
			req:  web_helpers.SignupBodyRequest{Username: "newuser", Password: "Admin123", Email: "new@example.com"},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
			},
			expectedErr: false,
		},
		{
			name: "ошибка репозитория",
			req:  web_helpers.SignupBodyRequest{Username: "newuser", Password: "Admin123", Email: "new@example.com"},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, errors.New("db error"))
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

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			authUser, err := uc.Create(context.Background(), c.req)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.req.Username, authUser.Username)
			}
		})
	}
}

func TestUserUseCase_GetByCredentials(t *testing.T) {
	t.Parallel()

	hashedUser := &models.UserModel{
		Id:       1,
		Username: "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
	}

	cases := []struct {
		name        string
		req         web_helpers.LoginBodyRequest
		setupMocks  func(repo *repomocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name: "пользователь не найден",
			req:  web_helpers.LoginBodyRequest{Username: "unknown", Password: "Admin123"},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByUsername(gomock.Any(), "unknown").Return(nil, errors.New("not found"))
			},
			expectedErr: true,
		},
		{
			name: "неверный пароль",
			req:  web_helpers.LoginBodyRequest{Username: "testuser", Password: "wrongpass"},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(hashedUser, nil)
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

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			_, err := uc.GetByCredentials(context.Background(), c.req)
			require.Error(t, err)
		})
	}
}

func TestUserUseCase_IsAuthUserExists(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{Id: 1, Username: "testuser", Email: "test@example.com"}

	cases := []struct {
		name       string
		isAuth     bool
		userId     int
		setupMocks func(repo *repomocks.MockUserRepository)
		expectOk   bool
	}{
		{
			name:       "не авторизован",
			isAuth:     false,
			userId:     1,
			setupMocks: func(repo *repomocks.MockUserRepository) {},
			expectOk:   false,
		},
		{
			name:   "авторизован, пользователь найден",
			isAuth: true,
			userId: 1,
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), 1).Return(testUser, nil)
			},
			expectOk: true,
		},
		{
			name:   "авторизован, пользователь не найден",
			isAuth: true,
			userId: 999,
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), 999).Return(nil, errors.New("not found"))
			},
			expectOk: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			_, ok := uc.IsAuthUserExists(context.Background(), c.isAuth, c.userId)
			require.Equal(t, c.expectOk, ok)
		})
	}
}

func TestUserUseCase_UpdateLastLogin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name: "успешно",
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().UpdateLastLogin(gomock.Any(), 1, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "ошибка",
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().UpdateLastLogin(gomock.Any(), 1, gomock.Any()).Return(errors.New("db error"))
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

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			err := uc.UpdateLastLogin(context.Background(), 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserUseCase_Update(t *testing.T) {
	t.Parallel()

	username := "newname"
	testUser := &models.UserModel{Id: 1, Username: username}

	cases := []struct {
		name        string
		profile     models.UpdateUserProfile
		setupMocks  func(repo *repomocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name:    "успешное обновление без пароля",
			profile: models.UpdateUserProfile{Id: 1, Username: &username, UpdatedAt: time.Now()},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(testUser, nil)
			},
			expectedErr: false,
		},
		{
			name:    "ошибка репозитория",
			profile: models.UpdateUserProfile{Id: 1, Username: &username, UpdatedAt: time.Now()},
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
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

			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo)

			uc := NewUser(repo)
			user, err := uc.Update(context.Background(), c.profile)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.Equal(t, username, user.Username)
			}
		})
	}
}
