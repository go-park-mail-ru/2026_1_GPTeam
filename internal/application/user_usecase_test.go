package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)


type dummyEnumsApp struct{}

func (d dummyEnumsApp) GetCurrencyCodes() []string    { return []string{"RUB", "USD", "EUR"} }
func (d dummyEnumsApp) GetTransactionTypes() []string { return []string{} }
func (d dummyEnumsApp) GetCategoryTypes() []string    { return []string{} }

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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
			_, err := uc.GetByCredentials(context.Background(), c.req)
			require.Error(t, err)
		})
	}
}

func TestUserUseCase_GetByCredentials_Success(t *testing.T) {
	t.Parallel()

	hashed, err := bcrypt.GenerateFromPassword([]byte("Admin123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	hashedUser := &models.UserModel{
		Id:       1,
		Username: "testuser",
		Password: string(hashed),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockUserRepository(ctrl)
	repo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(hashedUser, nil)

	uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
	user, err := uc.GetByCredentials(context.Background(), web_helpers.LoginBodyRequest{
		Username: "testuser",
		Password: "Admin123",
	})

	require.NoError(t, err)
	require.Equal(t, hashedUser.Username, user.Username)
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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
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

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
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

func TestUserUseCase_Update_WithPassword(t *testing.T) {
	t.Parallel()

	username := "user"
	testUser := &models.UserModel{Id: 1, Username: username}

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockUserRepository, capturedProfile *models.UpdateUserProfile)
		expectedErr bool
	}{
		{
			name: "успешное обновление с паролем",
			setupMocks: func(repo *repomocks.MockUserRepository, capturedProfile *models.UpdateUserProfile) {
				repo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p models.UpdateUserProfile) (*models.UserModel, error) {
						*capturedProfile = p
						return testUser, nil
					})
			},
			expectedErr: false,
		},
		{
			name: "ошибка репозитория при обновлении с паролем",
			setupMocks: func(repo *repomocks.MockUserRepository, capturedProfile *models.UpdateUserProfile) {
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

			originalPassword := "NewPass123"
			password := originalPassword
			profile := models.UpdateUserProfile{
				Id:        1,
				Username:  &username,
				Password:  &password,
				UpdatedAt: time.Now(),
			}

			var capturedProfile models.UpdateUserProfile
			repo := repomocks.NewMockUserRepository(ctrl)
			c.setupMocks(repo, &capturedProfile)

			uc := NewUser(repo, dummyEnumsApp{}, mocks.NewMockAvatarUploader(gomock.NewController(t)))
			user, err := uc.Update(context.Background(), profile)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, capturedProfile.Password)
				require.NotEqual(t, originalPassword, *capturedProfile.Password)
				err = bcrypt.CompareHashAndPassword([]byte(*capturedProfile.Password), []byte(originalPassword))
				require.NoError(t, err)
			}
		})
	}
}

func TestUserUseCase_UploadAvatar(t *testing.T) {
	cases := []struct {
		name        string
		setupRepo   func(repo *repomocks.MockUserRepository)
		setupAvatar func(av *mocks.MockAvatarUploader)
		fileContent string
		expectedErr bool
		expectedURL string
	}{
		{
			name:        "успешная загрузка",
			fileContent: "fake image data",
			setupAvatar: func(av *mocks.MockAvatarUploader) {
				av.EXPECT().Upload(gomock.Any(), gomock.Any(), ".png").Return("avatar.png", nil)
			},
			setupRepo: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().UpdateAvatar(gomock.Any(), 1, "avatar.png").Return(nil)
			},
			expectedErr: false,
			expectedURL: "avatar.png",
		},
		{
			name:        "ошибка хранилища",
			fileContent: "fake image data",
			setupAvatar: func(av *mocks.MockAvatarUploader) {
				av.EXPECT().Upload(gomock.Any(), gomock.Any(), ".png").Return("", errors.New("io error"))
			},
			setupRepo:   func(repo *repomocks.MockUserRepository) {},
			expectedErr: true,
		},
		{
			name:        "ошибка репозитория",
			fileContent: "fake image data",
			setupAvatar: func(av *mocks.MockAvatarUploader) {
				av.EXPECT().Upload(gomock.Any(), gomock.Any(), ".png").Return("avatar.png", nil)
			},
			setupRepo: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().UpdateAvatar(gomock.Any(), 1, "avatar.png").Return(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockUserRepository(ctrl)
			avatar := mocks.NewMockAvatarUploader(ctrl)
			c.setupRepo(repo)
			c.setupAvatar(avatar)

			uc := NewUser(repo, dummyEnumsApp{}, avatar)
			avatarURL, err := uc.UploadAvatar(
				context.Background(),
				1,
				strings.NewReader(c.fileContent),
				".png",
			)

			if c.expectedErr {
				require.Error(t, err)
				require.Empty(t, avatarURL)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedURL, avatarURL)
			}
		})
	}
}

func TestUser_IsStaff(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockUserRepository)
		userId     int
		isStaff    bool
		err        error
	}{
		{
			name: "staff",
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.UserModel{Id: 1, IsStaff: true}, nil)
			},
			userId:  1,
			isStaff: true,
			err:     nil,
		},
		{
			name: "no staff",
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.UserModel{Id: 2, IsStaff: false}, nil)
			},
			userId:  2,
			isStaff: false,
			err:     nil,
		},
		{
			name: "fail",
			setupMocks: func(repo *repomocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, repository.NothingInTableError)
			},
			userId:  3,
			isStaff: false,
			err:     repository.NothingInTableError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockUserRepository(ctrl)
			testCase.setupMocks(repo)
			enumsApp := mocks.NewMockEnumsUseCase(ctrl)
			app := NewUser(repo, enumsApp, mocks.NewMockAvatarUploader(gomock.NewController(t)))
			isStaff, err := app.IsStaff(context.Background(), testCase.userId)
			require.ErrorIs(t, testCase.err, err)
			require.Equal(t, testCase.isStaff, isStaff)
		})
	}
}
