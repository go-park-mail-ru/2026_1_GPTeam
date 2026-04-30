package application

import (
	"context"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSupport_Create(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockSupportRepository)
		data       web_helpers.SupportRequest
		userId     int
		res        int
		err        error
	}{
		{
			name: "ok",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
			},
			data: web_helpers.SupportRequest{
				Category: "a",
				Message:  "text",
			},
			userId: 1,
			res:    1,
			err:    nil,
		},
		{
			name: "fail",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, repository.ConstraintError)
			},
			data: web_helpers.SupportRequest{
				Category: strings.Repeat("a", 300),
				Message:  "text",
			},
			userId: 1,
			res:    -1,
			err:    repository.ConstraintError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockSupportRepository(ctrl)
			testCase.setupMocks(repo)
			app := NewSupport(repo)
			id, err := app.Create(context.Background(), testCase.data, testCase.userId)
			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.res, id)
		})
	}
}

func TestSupport_GetById(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockSupportRepository)
		id         int
		res        models.SupportModel
		err        error
	}{
		{
			name: "ok",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{Id: 1}, nil)
			},
			id:  1,
			res: models.SupportModel{Id: 1},
			err: nil,
		},
		{
			name: "fail",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(models.SupportModel{}, repository.ConstraintError)
			},
			id:  -1,
			res: models.SupportModel{},
			err: repository.ConstraintError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockSupportRepository(ctrl)
			testCase.setupMocks(repo)
			app := NewSupport(repo)
			support, err := app.GetById(context.Background(), testCase.id)
			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.res, support)
		})
	}
}

func TestSupport_GetAll(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockSupportRepository)
		res        []models.SupportModel
		err        error
	}{
		{
			name: "ok",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetAll(gomock.Any()).Return([]models.SupportModel{{Id: 1}, {Id: 2}}, nil)
			},
			res: []models.SupportModel{
				{Id: 1},
				{Id: 2},
			},
			err: nil,
		},
		{
			name: "empty",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetAll(gomock.Any()).Return([]models.SupportModel{}, repository.InvalidDataInTableError)
			},
			res: []models.SupportModel{},
			err: repository.InvalidDataInTableError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockSupportRepository(ctrl)
			testCase.setupMocks(repo)
			app := NewSupport(repo)
			supports, err := app.GetAll(context.Background())
			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.res, supports)
		})
	}
}

func TestSupport_GetAllByUser(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockSupportRepository)
		userId     int
		res        []models.SupportModel
		err        error
	}{
		{
			name: "ok",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetAllByUser(gomock.Any(), gomock.Any()).Return([]models.SupportModel{{Id: 2}}, nil)
			},
			userId: 2,
			res: []models.SupportModel{
				{Id: 2},
			},
			err: nil,
		},
		{
			name: "empty",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().GetAllByUser(gomock.Any(), gomock.Any()).Return([]models.SupportModel{}, nil)
			},
			userId: 1,
			res:    []models.SupportModel{},
			err:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockSupportRepository(ctrl)
			testCase.setupMocks(repo)
			app := NewSupport(repo)
			supports, err := app.GetAllByUser(context.Background(), testCase.userId)
			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.res, supports)
		})
	}
}

func TestSupport_Update(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(repo *repomocks.MockSupportRepository)
		id         int
		status     string
		err        error
	}{
		{
			name: "ok",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			id:  2,
			err: nil,
		},
		{
			name: "fail",
			setupMocks: func(repo *repomocks.MockSupportRepository) {
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(repository.NothingInTableError)
			},
			id:  1,
			err: repository.NothingInTableError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := repomocks.NewMockSupportRepository(ctrl)
			testCase.setupMocks(repo)
			app := NewSupport(repo)
			err := app.Update(context.Background(), testCase.id, testCase.status)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}
