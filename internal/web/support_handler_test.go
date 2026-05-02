package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSupportHandler_Create(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(supportApp *mocks.MockSupportUseCase)
		ctx   context.Context
		body  web_helpers.SupportRequest
		code  int
	}{
		{
			name: "ok",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(1, nil)
			},
			ctx: context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			body: web_helpers.SupportRequest{
				Category: "a",
				Message:  "b",
			},
			code: http.StatusOK,
		},
		{
			name:  "unauthorized",
			setup: func(supportApp *mocks.MockSupportUseCase) {},
			ctx:   context.Background(),
			body: web_helpers.SupportRequest{
				Category: "a",
				Message:  "b",
			},
			code: http.StatusUnauthorized,
		},
		{
			name: "fail (constraint error)",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(-1, repository.ConstraintError)
			},
			ctx: context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			body: web_helpers.SupportRequest{
				Category: "a",
				Message:  "b",
			},
			code: http.StatusBadRequest,
		},
		{
			name: "fail (duplicated error)",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(-1, repository.DuplicatedDataError)
			},
			ctx: context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			body: web_helpers.SupportRequest{
				Category: "a",
				Message:  "b",
			},
			code: http.StatusBadRequest,
		},
		{
			name: "fail (server error)",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(-1, errors.New("some error"))
			},
			ctx: context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			body: web_helpers.SupportRequest{
				Category: "a",
				Message:  "b",
			},
			code: http.StatusInternalServerError,
		},
		{
			name:  "validation fail",
			setup: func(supportApp *mocks.MockSupportUseCase) {},
			ctx:   context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			body: web_helpers.SupportRequest{
				Category: strings.Repeat("a", 300),
				Message:  "b",
			},
			code: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			supportApp := mocks.NewMockSupportUseCase(ctrl)
			userApp := mocks.NewMockUserUseCase(ctrl)
			testCase.setup(supportApp)
			handler := NewSupportHandler(supportApp, userApp)
			body, _ := json.Marshal(testCase.body)
			r := httptest.NewRequest(http.MethodPost, "/support/create_appeal", bytes.NewReader(body)).WithContext(testCase.ctx)
			w := httptest.NewRecorder()
			handler.Create(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
		})
	}
}

func TestSupportHandler_GetAll(t *testing.T) {
	allSupports := []models.SupportModel{
		{
			Id:       1,
			UserId:   5,
			Category: "a",
			Message:  "aaa",
			Status:   "OPEN",
		},
		{
			Id:       2,
			UserId:   12,
			Category: "b",
			Message:  "bbb",
			Status:   "IN_WORK",
		},
	}
	testCases := []struct {
		name  string
		setup func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase)
		ctx   context.Context
		code  int
	}{
		{
			name: "ok",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetAll(gomock.Any()).Return(allSupports, nil)
				for i := 0; i < len(allSupports); i++ {
					userApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(&models.UserModel{Id: allSupports[i].UserId}, nil)
				}
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusOK,
		},
		{
			name: "fail (get supports)",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetAll(gomock.Any()).Return([]models.SupportModel{}, errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusInternalServerError,
		},
		{
			name: "fail (get user)",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetAll(gomock.Any()).Return(allSupports, nil)
				userApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			supportApp := mocks.NewMockSupportUseCase(ctrl)
			userApp := mocks.NewMockUserUseCase(ctrl)
			testCase.setup(supportApp, userApp)
			handler := NewSupportHandler(supportApp, userApp)
			r := httptest.NewRequest(http.MethodGet, "/support/get_all_appeals", nil).WithContext(testCase.ctx)
			w := httptest.NewRecorder()
			handler.GetAll(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
		})
	}
}

func TestSupportHandler_Detail(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase)
		ctx   context.Context
		id    string
		code  int
	}{
		{
			name: "ok",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{Id: 1}, nil)
				userApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(&models.UserModel{Id: 1}, nil)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusOK,
		},
		{
			name:  "wrong id",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {},
			ctx:   context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:    "",
			code:  http.StatusBadRequest,
		},
		{
			name:  "id not int",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {},
			ctx:   context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:    "abc",
			code:  http.StatusBadRequest,
		},
		{
			name: "not found",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{}, repository.NothingInTableError)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusNotFound,
		},
		{
			name: "server error",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{}, errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusInternalServerError,
		},
		{
			name: "not found user",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{Id: 1}, nil)
				userApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, repository.NothingInTableError)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusNotFound,
		},
		{
			name: "server error (get user)",
			setup: func(supportApp *mocks.MockSupportUseCase, userApp *mocks.MockUserUseCase) {
				supportApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{Id: 1}, nil)
				userApp.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			supportApp := mocks.NewMockSupportUseCase(ctrl)
			userApp := mocks.NewMockUserUseCase(ctrl)
			testCase.setup(supportApp, userApp)
			handler := NewSupportHandler(supportApp, userApp)
			url := fmt.Sprintf("/support/get_all_appeals/%s", testCase.id)
			r := httptest.NewRequest(http.MethodGet, url, nil).WithContext(testCase.ctx)
			r.SetPathValue("id", testCase.id)
			w := httptest.NewRecorder()
			handler.Detail(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
		})
	}
}

func TestSupportHandler_GetAllByUser(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(supportApp *mocks.MockSupportUseCase)
		ctx   context.Context
		code  int
		data  web_helpers.SupportsResponse
	}{
		{
			name: "ok",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().GetAllByUser(gomock.Any(), gomock.Any()).Return([]models.SupportModel{{Id: 1}, {Id: 2}}, nil)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusOK,
			data: web_helpers.SupportsResponse{
				SimpleResponse: web_helpers.SimpleResponse{
					Code:    http.StatusOK,
					Message: "OK",
				},
				Supports: []web_helpers.SupportResponse{{Id: 1}, {Id: 2}},
			},
		},
		{
			name: "empty",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().GetAllByUser(gomock.Any(), gomock.Any()).Return([]models.SupportModel{}, nil)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusOK,
			data: web_helpers.SupportsResponse{
				SimpleResponse: web_helpers.SimpleResponse{
					Code:    http.StatusOK,
					Message: "OK",
				},
				Supports: nil,
			},
		},
		{
			name:  "unauthorized",
			setup: func(supportApp *mocks.MockSupportUseCase) {},
			ctx:   context.Background(),
			code:  http.StatusUnauthorized,
		},
		{
			name: "server error",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().GetAllByUser(gomock.Any(), gomock.Any()).Return([]models.SupportModel{}, errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			code: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			supportApp := mocks.NewMockSupportUseCase(ctrl)
			userApp := mocks.NewMockUserUseCase(ctrl)
			testCase.setup(supportApp)
			handler := NewSupportHandler(supportApp, userApp)
			r := httptest.NewRequest(http.MethodGet, "/support/get_appeals", nil).WithContext(testCase.ctx)
			w := httptest.NewRecorder()
			handler.GetAllByUser(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
			if testCase.code == http.StatusOK {
				body := json.NewDecoder(w.Result().Body)
				var result web_helpers.SupportsResponse
				err := body.Decode(&result)
				require.NoError(t, err)
				require.Equal(t, testCase.data, result)
			}
		})
	}
}

func TestSupportHandler_Update(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(supportApp *mocks.MockSupportUseCase)
		ctx   context.Context
		id    string
		code  int
		body  web_helpers.UpdateSupportStatusRequest
	}{
		{
			name: "ok",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusOK,
			body: web_helpers.UpdateSupportStatusRequest{Status: "a"},
		},
		{
			name:  "id not int",
			setup: func(supportApp *mocks.MockSupportUseCase) {},
			ctx:   context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:    "abc",
			code:  http.StatusBadRequest,
			body:  web_helpers.UpdateSupportStatusRequest{Status: "a"},
		},
		{
			name: "not found",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(repository.NothingInTableError)
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "10",
			code: http.StatusNotFound,
			body: web_helpers.UpdateSupportStatusRequest{Status: "a"},
		},
		{
			name: "server error",
			setup: func(supportApp *mocks.MockSupportUseCase) {
				supportApp.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			ctx:  context.WithValue(context.Background(), "user", models.UserModel{Id: 1}),
			id:   "1",
			code: http.StatusInternalServerError,
			body: web_helpers.UpdateSupportStatusRequest{Status: "a"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			supportApp := mocks.NewMockSupportUseCase(ctrl)
			userApp := mocks.NewMockUserUseCase(ctrl)
			testCase.setup(supportApp)
			handler := NewSupportHandler(supportApp, userApp)
			url := fmt.Sprintf("/support/update/%s", testCase.id)
			data, _ := json.Marshal(testCase.body)
			r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(data)).WithContext(testCase.ctx)
			r.SetPathValue("id", testCase.id)
			w := httptest.NewRecorder()
			handler.Update(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
		})
	}
}
