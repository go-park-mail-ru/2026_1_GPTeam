package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func ptr(s string) *string {
	return &s
}

func TestUserHandler_Profile(t *testing.T) {
	t.Parallel()

	testUserVal := &models.UserModel{Id: 1, Username: "testuser", Email: "test@test.com"}

	cases := []struct {
		name         string
		ctx          context.Context
		expectedCode int
	}{
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "success",
			ctx:          context.WithValue(context.Background(), "user", testUserVal),
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userApp := appmocks.NewMockUserUseCase(ctrl)
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			handler := NewUserHandler(userApp, accountApp)

			req := httptest.NewRequest(http.MethodGet, "/profile", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.ProfileHandler(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	testUserVal := &models.UserModel{Id: 1, Username: "testuser", Email: "test@test.com"}

	cases := []struct {
		name         string
		body         any
		ctx          context.Context
		setupMocks   func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			body:         web_helpers.UpdateUserProfileRequest{},
			ctx:          context.Background(),
			setupMocks:   func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "validation error",
			body:         web_helpers.UpdateUserProfileRequest{},
			ctx:          context.WithValue(context.Background(), "user", testUserVal),
			setupMocks:   func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "success",
			body: web_helpers.UpdateUserProfileRequest{Username: ptr("newname")},
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&models.UserModel{Id: testUserVal.Id, Username: "newname", Email: testUserVal.Email}, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userApp := appmocks.NewMockUserUseCase(ctrl)
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(userApp, accountApp)
			handler := NewUserHandler(userApp, accountApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPatch, "/profile", bytes.NewReader(bodyBytes)).WithContext(c.ctx)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ProfileHandler(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_Balance(t *testing.T) {
	t.Parallel()

	testUserVal := &models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			setupMocks:   func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "error",
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				accountApp.EXPECT().GetAllAccountsByUserIdWithBalance(gomock.Any(), testUserVal.Id).Return(nil, nil, nil, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				accounts := []models.AccountModel{{Currency: "RUB", Balance: 100}}
				incomes := []float64{50}
				expenses := []float64{25}
				accountApp.EXPECT().GetAllAccountsByUserIdWithBalance(gomock.Any(), testUserVal.Id).Return(accounts, incomes, expenses, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userApp := appmocks.NewMockUserUseCase(ctrl)
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(userApp, accountApp)
			handler := NewUserHandler(userApp, accountApp)

			req := httptest.NewRequest(http.MethodGet, "/balance", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.Balance(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_UploadAvatar(t *testing.T) {
	t.Parallel()

	testUserPtr := &models.UserModel{Id: 1, Username: "testuser"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userApp := appmocks.NewMockUserUseCase(ctrl)
		accountApp := appmocks.NewMockAccountUseCase(ctrl)
		userApp.EXPECT().UploadAvatar(gomock.Any(), testUserPtr.Id, gomock.Any(), gomock.Any()).Return("avatar.png", nil)

		handler := NewUserHandler(userApp, accountApp)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.png")
		part.Write([]byte("\x89PNG\x0D\x0A\x1A\x0A" + "fake content that makes file bigger than 512 bytes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                "))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/avatar", body).WithContext(context.WithValue(context.Background(), "user", testUserPtr))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		handler.UploadAvatar(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no file", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userApp := appmocks.NewMockUserUseCase(ctrl)
		accountApp := appmocks.NewMockAccountUseCase(ctrl)
		handler := NewUserHandler(userApp, accountApp)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/avatar", body).WithContext(context.WithValue(context.Background(), "user", testUserPtr))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		handler.UploadAvatar(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_IsStaff(t *testing.T) {
	staffUser := models.UserModel{Id: 1, Username: "admin", IsStaff: true}
	testUser := models.UserModel{Id: 2, Username: "test", IsStaff: false}

	testCases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(userApp *appmocks.MockUserUseCase)
		expectedCode int
		isStaff      bool
	}{
		{
			name: "staff",
			ctx:  context.WithValue(context.Background(), "user", staffUser),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			expectedCode: http.StatusOK,
			isStaff:      true,
		},
		{
			name: "not staff",
			ctx:  context.WithValue(context.Background(), "user", testUser),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			expectedCode: http.StatusOK,
			isStaff:      false,
		},
		{
			name: "error",
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{}),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(false, repository.NothingInTableError)
			},
			expectedCode: http.StatusInternalServerError,
			isStaff:      false,
		},
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			setupMocks:   func(userApp *appmocks.MockUserUseCase) {},
			expectedCode: http.StatusUnauthorized,
			isStaff:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userApp := appmocks.NewMockUserUseCase(ctrl)
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			testCase.setupMocks(userApp)
			handler := NewUserHandler(userApp, accountApp)
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			writer.Close()
			r := httptest.NewRequest(http.MethodGet, "/api/is_staff", nil).WithContext(testCase.ctx)
			w := httptest.NewRecorder()
			handler.IsStaff(w, r)
			require.Equal(t, testCase.expectedCode, w.Code)
			var response struct {
				IsStaff bool `json:"is_staff"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			require.Equal(t, testCase.isStaff, response.IsStaff)
		})
	}
}
