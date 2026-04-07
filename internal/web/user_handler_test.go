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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func ptr(s string) *string {
	return &s
}

func TestUserHandler_Profile(t *testing.T) {
	t.Parallel()

	testUserVal := models.UserModel{Id: 1, Username: "testuser", Email: "test@test.com"}

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
			handler := NewUserHandler(userApp)

			req := httptest.NewRequest(http.MethodGet, "/profile", nil).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.ProfileHandler(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	testUserVal := models.UserModel{Id: 1, Username: "testuser", Email: "test@test.com"}

	cases := []struct {
		name         string
		body         any
		ctx          context.Context
		setupMocks   func(userApp *appmocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			body:         web_helpers.UpdateUserProfileRequest{},
			ctx:          context.Background(),
			setupMocks:   func(userApp *appmocks.MockUserUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "validation error",
			body:         web_helpers.UpdateUserProfileRequest{},
			ctx:          context.WithValue(context.Background(), "user", testUserVal),
			setupMocks:   func(userApp *appmocks.MockUserUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "success",
			body: web_helpers.UpdateUserProfileRequest{Username: ptr("newname")},
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&models.UserModel{Username: "newname"}, nil)
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
			c.setupMocks(userApp)
			handler := NewUserHandler(userApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPatch, "/profile", bytes.NewReader(bodyBytes)).WithContext(c.ctx)
			w := httptest.NewRecorder()
			handler.ProfileHandler(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_Balance(t *testing.T) {
	t.Parallel()

	testUserVal := models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		ctx          context.Context
		setupMocks   func(userApp *appmocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name:         "unauthorized",
			ctx:          context.Background(),
			setupMocks:   func(userApp *appmocks.MockUserUseCase) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "error",
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().GetUserBalance(gomock.Any(), testUserVal.Id).Return(nil, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user", testUserVal),
			setupMocks: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().GetUserBalance(gomock.Any(), testUserVal.Id).Return([]models.CurrencyStat{{Currency: "RUB", Balance: 100}}, nil)
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
			c.setupMocks(userApp)
			handler := NewUserHandler(userApp)

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
		userApp.EXPECT().UploadAvatar(gomock.Any(), testUserPtr.Id, gomock.Any(), gomock.Any()).Return("avatar.png", nil)

		handler := NewUserHandler(userApp)

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
		handler := NewUserHandler(userApp)

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
