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

func newAccountPostgres(t *testing.T) (*AccountPostgres, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return &AccountPostgres{db: mock}, mock
}

func TestAccountPostgres_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()
	account := models.AccountModel{Name: "base", Balance: 0, Currency: "RUB", CreatedAt: now, UpdatedAt: now}

	cases := []struct {
		name        string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedId  int
		expectedErr bool
	}{
		{
			name: "успешное создание",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`insert into account`).
					WithArgs("base", float64(0), "RUB", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedId:  1,
			expectedErr: false,
		},
		{
			name: "ошибка БД",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into account`).
					WithArgs("base", float64(0), "RUB", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			expectedId:  -1,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newAccountPostgres(t)
			c.setupMock(mock)

			id, err := repo.Create(context.Background(), account)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedId, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAccountPostgres_LinkAccountAndUser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedErr bool
	}{
		{
			name: "успешная линковка",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`insert into account_user`).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name: "ошибка БД",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into account_user`).
					WithArgs(1, 1).
					WillReturnError(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newAccountPostgres(t)
			c.setupMock(mock)

			id, err := repo.LinkAccountAndUser(context.Background(), 1, 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAccountPostgres_GetIdsByUserAndAccount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedLen int
		expectedErr bool
	}{
		{
			name: "найдены записи",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
				mock.ExpectQuery(`select id from account_user`).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			expectedLen: 2,
			expectedErr: false,
		},
		{
			name: "ошибка БД",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id from account_user`).
					WithArgs(1, 1).
					WillReturnError(errors.New("db error"))
			},
			expectedLen: 0,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newAccountPostgres(t)
			c.setupMock(mock)

			ids, err := repo.GetIdsByUserAndAccount(context.Background(), 1, 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, ids, c.expectedLen)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAccountPostgres_GetAccountIdByUserId(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectedId  int
		expectedErr bool
	}{
		{
			name: "найден",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"account_id"}).AddRow(42)
				mock.ExpectQuery(`SELECT account_id FROM account_user`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedId:  42,
			expectedErr: false,
		},
		{
			name: "не найден",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT account_id FROM account_user`).
					WithArgs(1).
					WillReturnError(errors.New("no rows"))
			},
			expectedId:  0,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			repo, mock := newAccountPostgres(t)
			c.setupMock(mock)

			id, err := repo.GetAccountIdByUserId(context.Background(), 1)

			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedId, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
