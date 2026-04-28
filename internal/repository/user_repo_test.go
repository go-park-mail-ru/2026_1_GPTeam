package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func TestUserPostgres_Create(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		user          models.UserModel
		setupMock     func(mock pgxmock.PgxPoolIface)
		expectedId    int
		expectedErr   bool
		expectedErrIs error
	}{
		{
			name: "успешное создание",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedId:  1,
			expectedErr: false,
		},
		{
			name: "ошибка БД",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			expectedId:  -1,
			expectedErr: true,
		},
		{
			name: "UniqueViolation — DuplicatedDataError",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			expectedId:    -1,
			expectedErr:   true,
			expectedErrIs: DuplicatedDataError,
		},
		{
			name: "CheckViolation — ConstraintError",
			user: models.UserModel{Username: "testuser", Password: "hash", Email: "test@example.com"},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into "user"`).
					WithArgs("testuser", "hash", "test@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			expectedId:    -1,
			expectedErr:   true,
			expectedErrIs: ConstraintError,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			id, err := repo.Create(context.Background(), c.user)

			if c.expectedErr {
				require.Error(t, err)
				require.Equal(t, c.expectedId, id)
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs)
				}
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
	lastLoginTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		name              string
		id                int
		setupMock         func(mock pgxmock.PgxPoolIface)
		expectedErr       bool
		expectedLastLogin *time.Time
	}{
		{
			name: "пользователь найден, last_login заполнен",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "password", "email",
					"created_at", "last_login", "avatar_url", "updated_at", "active", "is_staff",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, lastLoginTime, "", now, true, false)
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedErr:       false,
			expectedLastLogin: &lastLoginTime,
		},
		{
			name: "пользователь не найден",
			id:   999,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where id = \$1`).
					WithArgs(999).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			user, err := repo.GetByID(context.Background(), c.id)

			if c.expectedErr {
				require.ErrorIs(t, err, NothingInTableError)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, c.id, user.Id)
				if c.expectedLastLogin != nil {
					require.Equal(t, *c.expectedLastLogin, user.LastLogin)
				}
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
					"created_at", "last_login", "avatar_url", "updated_at", "active", "is_staff",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, nil, "", now, true, false)
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where username = \$1`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:     "не найден",
			username: "unknown",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where username = \$1`).
					WithArgs("unknown").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			user, err := repo.GetByUsername(context.Background(), c.username)

			if c.expectedErr {
				require.ErrorIs(t, err, NothingInTableError)
				require.Nil(t, user)
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
					"created_at", "last_login", "avatar_url", "updated_at", "active", "is_staff",
				}).AddRow(1, "testuser", "hash", "test@example.com", now, nil, "", now, true, false)
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:  "не найден",
			email: "unknown@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where email = \$1`).
					WithArgs("unknown@example.com").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			user, err := repo.GetByEmail(context.Background(), c.email)

			if c.expectedErr {
				require.ErrorIs(t, err, NothingInTableError)
				require.Nil(t, user)
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
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			err = repo.UpdateLastLogin(context.Background(), 1, time.Now())

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
		name          string
		profile       models.UpdateUserProfile
		setupMock     func(mock pgxmock.PgxPoolIface)
		expectedErr   bool
		expectedErrIs error
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
						pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
						pgxmock.AnyArg(), pgxmock.AnyArg(), 1,
					).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:    "пользователь не найден",
			profile: models.UpdateUserProfile{Id: 999, Username: &username, UpdatedAt: now},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`UPDATE\s+"user"\s+SET`).
					WithArgs(
						pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
						pgxmock.AnyArg(), pgxmock.AnyArg(), 999,
					).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr:   true,
			expectedErrIs: NothingInTableError,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			user, err := repo.Update(context.Background(), c.profile)

			if c.expectedErr {
				require.Error(t, err)
				require.Nil(t, user)
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, username, user.Username)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_UpdateAvatar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		id            int
		avatarUrl     string
		setupMock     func(mock pgxmock.PgxPoolIface)
		expectedErr   bool
		expectedErrIs error
	}{
		{
			name:      "успешное обновление",
			id:        1,
			avatarUrl: "https://example.com/avatar.jpg",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`update "user" set avatar_url`).
					WithArgs("https://example.com/avatar.jpg", 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: false,
		},
		{
			name:      "пользователь не найден (RowsAffected = 0)",
			id:        999,
			avatarUrl: "https://example.com/avatar.jpg",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`update "user" set avatar_url`).
					WithArgs("https://example.com/avatar.jpg", 999).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErr:   true,
			expectedErrIs: NothingInTableError,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			repo := NewUserPostgres(mock)
			c.setupMock(mock)

			err = repo.UpdateAvatar(context.Background(), c.id, c.avatarUrl)

			if c.expectedErr {
				require.Error(t, err)
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs)
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
