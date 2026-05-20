package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=account_user.go -destination=mocks/mock_account_user.go -package=mocks
type AccountUserRepository interface {
	SearchUsers(ctx context.Context, accountId int, query string, limit int) ([]models.UserSearchResult, error)
	CreateInvite(ctx context.Context, accountId int, userId int) (models.AccountUserModel, error)
	GetByAccountIdAndUserId(ctx context.Context, accountId int, userId int) (models.AccountUserModel, error)
	GetMembersByAccountId(ctx context.Context, accountId int) ([]models.MemberResponse, error)
	UpdateStatus(ctx context.Context, accountId int, userId int, status string) (models.AccountUserModel, error)
	DeleteMember(ctx context.Context, accountId int, userId int) error
	GetOwnerByAccountId(ctx context.Context, accountId int) (int, error)
	GetPendingInvitesByUserId(ctx context.Context, userId int) ([]models.PendingInviteView, error)
	LeaveAccount(ctx context.Context, accountId int, userId int) error
}

type AccountUserPostgres struct {
	db DB
}

func NewAccountUserPostgres(db DB) *AccountUserPostgres {
	return &AccountUserPostgres{
		db: db,
	}
}

func mapAccountUserPgError(ctx context.Context, err error, action string) error {
	if err == nil {
		return nil
	}
	log := logger.GetLoggerWithRequestId(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNothingInTable
	}
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error(action, zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return ErrAccountDuplicatedData
		case pgerrcode.CheckViolation:
			return ErrConstraint
		case pgerrcode.ForeignKeyViolation:
			return ErrAccountForeignKey
		default:
			return pgErr
		}
	}
	log.Error(action, zap.Error(err))
	return err
}

func (obj *AccountUserPostgres) SearchUsers(ctx context.Context, accountId int, query string, limit int) ([]models.UserSearchResult, error) {
	queryStr := `
        SELECT u.id, u.username
        FROM "user" u
        WHERE (u.username ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')
          AND u.id != COALESCE((SELECT owner_id FROM account WHERE id = $3), 0)
          AND NOT EXISTS (
              SELECT 1 FROM account_user au
              WHERE au.account_id = $3
                AND au.user_id = u.id
                AND au.deleted_at IS NULL
          )
        LIMIT $2`

	args := []any{query, limit, accountId}

	rows, err := obj.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.UserSearchResult
	for rows.Next() {
		var res models.UserSearchResult
		if err := rows.Scan(&res.Id, &res.Username); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

// CreateInvite создаёт приглашение. Если пользователь был ранее кикнут (deleted_at IS NOT NULL),
// строка сбрасывается и возвращается заново. Если пользователь активен — RETURNING вернёт 0 строк
// → ErrNothingInTable, usecase смаппит это в ErrInviteAlreadyExists / ErrAlreadyMember.
func (obj *AccountUserPostgres) CreateInvite(ctx context.Context, accountId int, userId int) (models.AccountUserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		INSERT INTO account_user (account_id, user_id, status, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (account_id, user_id) DO UPDATE
			SET status        = 'pending',
			    deleted_at    = NULL,
			    deleted_reason = NULL,
			    created_at    = now()
			WHERE account_user.deleted_at IS NOT NULL
		RETURNING id, account_id, user_id, status, created_at`
	args := []any{accountId, userId, AccountUserStatusPending}

	var accountUser models.AccountUserModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&accountUser.Id,
		&accountUser.AccountId,
		&accountUser.UserId,
		&accountUser.Status,
		&accountUser.CreatedAt,
	)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if mappedErr := mapAccountUserPgError(ctx, err, "failed to create invite"); mappedErr != nil {
		return models.AccountUserModel{}, mappedErr
	}

	log.Info("Query executed")
	return accountUser, nil
}

func (obj *AccountUserPostgres) GetByAccountIdAndUserId(ctx context.Context, accountId int, userId int) (models.AccountUserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		SELECT id, account_id, user_id, status, created_at
		FROM account_user
		WHERE account_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	args := []any{accountId, userId}

	var accountUser models.AccountUserModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&accountUser.Id,
		&accountUser.AccountId,
		&accountUser.UserId,
		&accountUser.Status,
		&accountUser.CreatedAt,
	)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if mappedErr := mapAccountUserPgError(ctx, err, "failed to get account user"); mappedErr != nil {
		return models.AccountUserModel{}, mappedErr
	}

	log.Info("Query executed")
	return accountUser, nil
}

func (obj *AccountUserPostgres) GetMembersByAccountId(ctx context.Context, accountId int) ([]models.MemberResponse, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		SELECT
			COALESCE(au.id, 0)                        AS id,
			a.id                                       AS account_id,
			u.id                                       AS user_id,
			u.username,
			u.email,
			$2::account_user_status                    AS status,
			COALESCE(au.created_at, a.created_at)     AS created_at,
			true                                       AS is_owner
		FROM account a
		JOIN "user" u ON u.id = a.owner_id
		LEFT JOIN account_user au
			ON au.account_id = a.id AND au.user_id = a.owner_id AND au.deleted_at IS NULL
		WHERE a.id = $1

		UNION ALL

		SELECT
			au.id,
			au.account_id,
			u.id,
			u.username,
			u.email,
			au.status,
			au.created_at,
			false
		FROM account_user au
		JOIN "user" u ON u.id = au.user_id
		JOIN account a ON a.id = au.account_id
		WHERE au.account_id = $1
		  AND au.user_id != a.owner_id
		  AND au.status = $2
		  AND au.deleted_at IS NULL

		ORDER BY is_owner DESC, created_at ASC`
	args := []any{accountId, AccountUserStatusAccepted}

	start := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if err != nil {
		log.Error("failed to get members by account id", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	members := make([]models.MemberResponse, 0)
	for rows.Next() {
		var member models.MemberResponse
		if err = rows.Scan(
			&member.Id,
			&member.AccountId,
			&member.UserId,
			&member.Username,
			&member.Email,
			&member.Status,
			&member.CreatedAt,
			&member.IsOwner,
		); err != nil {
			log.Error("failed to scan member", zap.Error(err))
			return nil, ErrInvalidDataInTable
		}
		members = append(members, member)
	}

	if rows.Err() != nil {
		log.Error("failed while reading members", zap.Error(rows.Err()))
		return nil, rows.Err()
	}

	log.Info("Query executed")
	return members, nil
}

func (obj *AccountUserPostgres) UpdateStatus(ctx context.Context, accountId int, userId int, status string) (models.AccountUserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		UPDATE account_user
		SET status = $1
		WHERE account_id = $2 AND user_id = $3 AND deleted_at IS NULL
		RETURNING id, account_id, user_id, status, created_at`
	args := []any{AccountUserStatus(status), accountId, userId}

	var accountUser models.AccountUserModel
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&accountUser.Id,
		&accountUser.AccountId,
		&accountUser.UserId,
		&accountUser.Status,
		&accountUser.CreatedAt,
	)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if mappedErr := mapAccountUserPgError(ctx, err, "failed to update account user status"); mappedErr != nil {
		return models.AccountUserModel{}, mappedErr
	}

	log.Info("Query executed")
	return accountUser, nil
}

func (obj *AccountUserPostgres) DeleteMember(ctx context.Context, accountId int, userId int) error {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		UPDATE account_user
		SET deleted_at = now(), deleted_reason = $3
		WHERE account_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	args := []any{accountId, userId, AccountUserDeletedReasonKicked}

	start := time.Now()
	result, err := obj.db.Exec(ctx, query, args...)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if err != nil {
		log.Error("failed to delete member", zap.Error(err))
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNothingInTable
	}

	log.Info("Query executed")
	return nil
}

func (obj *AccountUserPostgres) GetOwnerByAccountId(ctx context.Context, accountId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		SELECT owner_id
		FROM account
		WHERE id = $1`
	args := []any{accountId}

	var ownerId int
	start := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&ownerId)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if mappedErr := mapAccountUserPgError(ctx, err, "failed to get owner by account id"); mappedErr != nil {
		return 0, mappedErr
	}

	log.Info("Query executed")
	return ownerId, nil
}

func (obj *AccountUserPostgres) GetPendingInvitesByUserId(ctx context.Context, userId int) ([]models.PendingInviteView, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		SELECT au.id, au.account_id, au.user_id, au.status, au.created_at, COALESCE(a.name, '')
		FROM account_user au
		INNER JOIN account a ON a.id = au.account_id AND a.deleted_at IS NULL
		WHERE au.user_id = $1
		  AND au.status = $2
		  AND au.user_id != a.owner_id
		  AND au.deleted_at IS NULL
		ORDER BY au.created_at DESC`
	args := []any{userId, AccountUserStatusPending}

	start := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if err != nil {
		log.Error("failed to get pending invites by user id", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	invites := make([]models.PendingInviteView, 0)
	for rows.Next() {
		var invite models.PendingInviteView
		if err = rows.Scan(
			&invite.Id,
			&invite.AccountId,
			&invite.UserId,
			&invite.Status,
			&invite.CreatedAt,
			&invite.AccountName,
		); err != nil {
			log.Error("failed to scan invite", zap.Error(err))
			return nil, ErrInvalidDataInTable
		}
		invites = append(invites, invite)
	}

	if rows.Err() != nil {
		log.Error("failed while reading invites", zap.Error(rows.Err()))
		return nil, rows.Err()
	}

	log.Info("Query executed")
	return invites, nil
}

// LeaveAccount помечает участника как вышедшего. Проверка что юзер не овнер — на уровне usecase.
func (obj *AccountUserPostgres) LeaveAccount(ctx context.Context, accountId int, userId int) error {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `
		UPDATE account_user
		SET deleted_at = now(), deleted_reason = $3
		WHERE account_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	start := time.Now()
	result, err := obj.db.Exec(ctx, query, accountId, userId, AccountUserDeletedReasonLeft)
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, []any{accountId, userId, AccountUserDeletedReasonLeft}, duration)
	metrics.GetMetrics().DbQueryDuration.WithLabelValues(query, "account_user").Observe(float64(duration.Milliseconds()))

	if err != nil {
		log.Error("failed to leave account", zap.Error(err))
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNothingInTable
	}

	log.Info("User left account", zap.Int("account_id", accountId), zap.Int("user_id", userId))
	return nil
}
