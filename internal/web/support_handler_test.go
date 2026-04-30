package web

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

		})
	}
}

func TestSupportHandler_GetAll(t *testing.T) {

}

func TestSupportHandler_Detail(t *testing.T) {

}

func TestSupportHandler_GetAllByUser(t *testing.T) {

}

func TestSupportHandler_Update(t *testing.T) {

}
