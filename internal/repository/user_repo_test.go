package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func newUserPostgres(t *testing.T) (*UserPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return &UserPostgres{db: mock}, mock
}

func TestUserPostgres_Create(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		user        models.UserModel
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedId  int
		expectedErr error
	}{
		{
			name: "успешное создание",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedId:  1,
			expectedErr: nil,
		},
		{
			name: "ошибка БД",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			expectedId:  -1,
			expectedErr: errors.New("db error"),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			id, err := repo.Create(context.Background(), c.user)

			if c.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedId, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()

	cases := []struct {
		name        string
		id          int
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name: "пользователь найден",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "password", "email",
					"created_at", "last_login", "avatar_url", "updated_at", "active",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, nil, "", now, true)
				mock.ExpectQuery(`select id, username`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name: "пользователь не найден",
			id:   999,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username`).
					WithArgs(999).
					WillReturnError(errors.New("no rows"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			user, err := repo.GetByID(context.Background(), c.id)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, c.id, user.Id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetByUsername(t *testing.T) {
	t.Parallel()

	now := time.Now()

	cases := []struct {
		name        string
		username    string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name:     "найден",
			username: "testuser",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "password", "email",
					"created_at", "last_login", "avatar_url", "updated_at", "active",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, nil, "", now, true)
				mock.ExpectQuery(`select id, username`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:     "не найден",
			username: "unknown",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username`).
					WithArgs("unknown").
					WillReturnError(errors.New("no rows"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			user, err := repo.GetByUsername(context.Background(), c.username)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.username, user.Username)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetByEmail(t *testing.T) {
	t.Parallel()

	now := time.Now()

	cases := []struct {
		name        string
		email       string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name:  "найден",
			email: "test@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "password", "email",
					"created_at", "last_login", "avatar_url", "updated_at", "active",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, nil, "", now, true)
				mock.ExpectQuery(`select id, username`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:  "не найден",
			email: "unknown@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username`).
					WithArgs("unknown@example.com").
					WillReturnError(errors.New("no rows"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			user, err := repo.GetByEmail(context.Background(), c.email)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.email, user.Email)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_UpdateLastLogin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name: "успешно",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`UPDATE "user" SET last_login`).
					WithArgs(pgxmock.AnyArg(), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: false,
		},
		{
			name: "ошибка",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`UPDATE "user" SET last_login`).
					WithArgs(pgxmock.AnyArg(), 1).
					WillReturnError(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			err := repo.UpdateLastLogin(context.Background(), 1, time.Now())

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()
	username := "newname"

	cases := []struct {
		name        string
		profile     models.UpdateUserProfile
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name:    "успешно",
			profile: models.UpdateUserProfile{Id: 1, Username: &username, UpdatedAt: now},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "password", "email",
					"created_at", "last_login", "avatar_url", "updated_at", "active",
				}).AddRow(1, username, "hash", "test@example.com", now, nil, "", now, true)
				mock.ExpectQuery(`UPDATE\s+"user"\s+SET`).
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						1,
					).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:    "ошибка",
			profile: models.UpdateUserProfile{Id: 1, Username: &username, UpdatedAt: now},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`UPDATE\s+"user"\s+SET`).
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						1,
					).
					WillReturnError(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newUserPostgres(t)
			c.setupMock(mock)

			user, err := repo.Update(context.Background(), c.profile)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.Equal(t, username, user.Username)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
