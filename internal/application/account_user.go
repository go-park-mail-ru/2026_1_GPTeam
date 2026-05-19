package application

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type AccountUserUseCase interface {
	SearchUsers(ctx context.Context, accountId int, query string, limit int) ([]models.UserSearchResult, error)
	CreateInvite(ctx context.Context, accountId int, ownerId int, targetUserId int) (models.AccountUserModel, error)
	GetMembers(ctx context.Context, accountId int, requestingUserId int) ([]models.MemberResponse, error)
	AcceptInvite(ctx context.Context, accountId int, userId int) (models.AccountUserModel, error)
	RejectInvite(ctx context.Context, accountId int, userId int) error
	RemoveMember(ctx context.Context, accountId int, ownerId int, targetUserId int) error
	GetPendingInvites(ctx context.Context, userId int) ([]models.PendingInviteView, error)
	LeaveAccount(ctx context.Context, accountId int, userId int) error
}

type AccountUserApp struct {
	accountUserRepo repository.AccountUserRepository
	accountRepo     repository.AccountRepository
}

func NewAccountUserApp(accountUserRepo repository.AccountUserRepository, accountRepo repository.AccountRepository) *AccountUserApp {
	return &AccountUserApp{
		accountUserRepo: accountUserRepo,
		accountRepo:     accountRepo,
	}
}

func (obj *AccountUserApp) SearchUsers(
	ctx context.Context,
	accountId int,
	query string,
	limit int,
) ([]models.UserSearchResult, error) {

	if limit <= 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	return obj.accountUserRepo.SearchUsers(ctx, accountId, query, limit)
}

func (obj *AccountUserApp) CreateInvite(
	ctx context.Context,
	accountId int,
	ownerId int,
	targetUserId int,
) (models.AccountUserModel, error) {

	if targetUserId == ownerId {
		return models.AccountUserModel{}, ErrSelfInvite
	}

	ownerIdFromDb, err := obj.accountUserRepo.GetOwnerByAccountId(ctx, accountId)
	if err != nil {
		return models.AccountUserModel{}, err
	}

	if ownerIdFromDb != ownerId {
		return models.AccountUserModel{}, ErrNotOwner
	}

	if ownerIdFromDb == targetUserId {
		return models.AccountUserModel{}, ErrSelfInvite
	}

	// Проверяем активную запись (deleted_at IS NULL).
	// Кикнутые юзеры (deleted_at IS NOT NULL) GetByAccountIdAndUserId не найдёт →
	// CreateInvite через UPSERT сбросит deleted_at и создаст новое приглашение.
	existing, err := obj.accountUserRepo.GetByAccountIdAndUserId(ctx, accountId, targetUserId)
	if err == nil {
		switch existing.Status {
		case string(models.InviteStatusPending):
			return models.AccountUserModel{}, ErrInviteAlreadyExists
		case string(models.InviteStatusAccepted):
			return models.AccountUserModel{}, ErrAlreadyMember
		}
	}

	result, err := obj.accountUserRepo.CreateInvite(ctx, accountId, targetUserId)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			// UPSERT вернул 0 строк — юзер активен, конфликт не был сброшен.
			return models.AccountUserModel{}, ErrInviteAlreadyExists
		}
		return models.AccountUserModel{}, err
	}
	return result, nil
}

func (obj *AccountUserApp) GetMembers(
	ctx context.Context,
	accountId int,
	requestingUserId int,
) ([]models.MemberResponse, error) {

	_, err := obj.accountRepo.GetById(ctx, requestingUserId, accountId)
	if err != nil {
		return nil, err
	}

	return obj.accountUserRepo.GetMembersByAccountId(ctx, accountId)
}

func (obj *AccountUserApp) AcceptInvite(
	ctx context.Context,
	accountId int,
	userId int,
) (models.AccountUserModel, error) {

	invite, err := obj.accountUserRepo.GetByAccountIdAndUserId(ctx, accountId, userId)
	if err != nil {
		return models.AccountUserModel{}, ErrInviteNotFound
	}

	if invite.Status != string(models.InviteStatusPending) {
		return models.AccountUserModel{}, errors.New("invite is not pending")
	}

	return obj.accountUserRepo.UpdateStatus(ctx, accountId, userId, string(models.InviteStatusAccepted))
}

func (obj *AccountUserApp) RejectInvite(
	ctx context.Context,
	accountId int,
	userId int,
) error {

	invite, err := obj.accountUserRepo.GetByAccountIdAndUserId(ctx, accountId, userId)
	if err != nil {
		return ErrInviteNotFound
	}

	if invite.Status != string(models.InviteStatusPending) {
		return errors.New("invite is not pending")
	}

	return obj.accountUserRepo.DeleteMember(ctx, accountId, userId)
}

func (obj *AccountUserApp) RemoveMember(
	ctx context.Context,
	accountId int,
	ownerId int,
	targetUserId int,
) error {

	ownerIdFromDb, err := obj.accountUserRepo.GetOwnerByAccountId(ctx, accountId)
	if err != nil {
		return err
	}

	if ownerIdFromDb != ownerId {
		return ErrNotOwner
	}

	if ownerIdFromDb == targetUserId {
		return ErrCannotRemoveOwner
	}

	err = obj.accountUserRepo.DeleteMember(ctx, accountId, targetUserId)
	if errors.Is(err, repository.NothingInTableError) {
		return ErrMemberNotFound
	}
	return err
}

func (obj *AccountUserApp) GetPendingInvites(
	ctx context.Context,
	userId int,
) ([]models.PendingInviteView, error) {

	return obj.accountUserRepo.GetPendingInvitesByUserId(ctx, userId)
}

func (obj *AccountUserApp) LeaveAccount(ctx context.Context, accountId int, userId int) error {
	ownerIdFromDb, err := obj.accountUserRepo.GetOwnerByAccountId(ctx, accountId)
	if err != nil {
		return err
	}
	if ownerIdFromDb == userId {
		return ErrOwnerCannotLeave
	}

	_, err = obj.accountUserRepo.GetByAccountIdAndUserId(ctx, accountId, userId)
	if err != nil {
		return err
	}

	return obj.accountUserRepo.LeaveAccount(ctx, accountId, userId)
}
