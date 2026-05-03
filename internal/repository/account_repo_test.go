package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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
					WithArgs("base", float64(0), "RUB", pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedId:  1,
			expectedErr: false,
		},
		{
			name: "ошибка БД",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`insert into account`).
					WithArgs("base", float64(0), "RUB", pgxmock.AnyArg()).
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

func TestAccountPostgres_GetAllAccountsByUserIdWithBalance(t *testing.T) {
	now := time.Now()

	query := `(?s)select.*a\.id.*a\.name.*a\.balance.*a\.currency.*a\.created_at.*a\.updated_at.*coalesce\(sum\(t\.value\).*as income.*coalesce\(sum\(t\.value\).*as expense.*from account a.*join account_user au.*left join transaction t.*where au\.user_id = \$1.*a\.deleted_at is null.*group by.*order by a\.id`

	testCases := []struct {
		name     string
		setup    func(mock pgxmock.PgxPoolIface)
		userId   int
		accounts []models.AccountModel
		income   []float64
		expense  []float64
		err      error
	}{
		{
			name: "ok",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"balance",
					"currency",
					"created_at",
					"updated_at",
					"income",
					"expense",
				}).
					AddRow(1, "a", 100.0, "RUB", now, now, 19.5, 3.0).
					AddRow(2, "b", 42.0, "RUB", now, now.Add(time.Hour), 27.0, 1.0)

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			userId: 1,
			accounts: []models.AccountModel{
				{
					Id:        1,
					Name:      "a",
					Balance:   100,
					Currency:  "RUB",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					Id:        2,
					Name:      "b",
					Balance:   42,
					Currency:  "RUB",
					CreatedAt: now,
					UpdatedAt: now.Add(time.Hour),
				},
			},
			income:  []float64{19.5, 27},
			expense: []float64{3, 1},
			err:     nil,
		},
		{
			name: "empty",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"balance",
					"currency",
					"created_at",
					"updated_at",
					"income",
					"expense",
				})

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			userId:   1,
			accounts: []models.AccountModel{},
			income:   []float64{},
			expense:  []float64{},
			err:      nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo, mock := newAccountPostgres(t)
			testCase.setup(mock)

			accounts, income, expense, err := repo.GetAllAccountsByUserIdWithBalance(context.Background(), testCase.userId)

			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.accounts, accounts)
			require.Equal(t, testCase.income, income)
			require.Equal(t, testCase.expense, expense)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAccountPostgres_GetAllAccountsByUserId(t *testing.T) {
	now := time.Now()

	query := `(?s)select.*a\.id.*a\.name.*a\.balance.*a\.currency.*a\.created_at.*a\.updated_at.*from account a.*join account_user au.*where au\.user_id = \$1.*a\.deleted_at is null.*order by a\.id`

	testCases := []struct {
		name     string
		setup    func(mock pgxmock.PgxPoolIface)
		userId   int
		accounts []models.AccountModel
		err      error
	}{
		{
			name: "ok",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"balance",
					"currency",
					"created_at",
					"updated_at",
				}).AddRow(1, "a", 100.0, "RUB", now, now)

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			userId: 1,
			accounts: []models.AccountModel{
				{
					Id:        1,
					Name:      "a",
					Balance:   100,
					Currency:  "RUB",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			err: nil,
		},
		{
			name: "empty",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"balance",
					"currency",
					"created_at",
					"updated_at",
				})

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			userId:   1,
			accounts: []models.AccountModel{},
			err:      nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo, mock := newAccountPostgres(t)
			testCase.setup(mock)

			accounts, err := repo.GetAllAccountsByUserId(context.Background(), testCase.userId)

			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.accounts, accounts)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAccountPostgres_GetByAccountId(t *testing.T) {
	now := time.Now()

	query := `(?s)select.*id.*name.*balance.*currency.*created_at.*updated_at.*from account.*where id = \$1.*deleted_at is null`

	testCases := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		id      int
		account models.AccountModel
		err     error
	}{
		{
			name: "ok",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"balance",
					"currency",
					"created_at",
					"updated_at",
				}).AddRow(1, "a", 100.0, "RUB", now, now)

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			id: 1,
			account: models.AccountModel{
				Id:        1,
				Name:      "a",
				Balance:   100,
				Currency:  "RUB",
				CreatedAt: now,
				UpdatedAt: now,
			},
			err: nil,
		},
		{
			name: "fail",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnError(pgx.ErrNoRows)
			},
			id:      1,
			account: models.AccountModel{},
			err:     NothingInTableError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo, mock := newAccountPostgres(t)
			testCase.setup(mock)

			account, err := repo.GetByAccountId(context.Background(), testCase.id)

			require.ErrorIs(t, err, testCase.err)
			require.Equal(t, testCase.account, account)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
