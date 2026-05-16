package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

func accountIDFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("accountId"))
}

func userIDFromMembersPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("userId"))
}

type AccountUserHandler struct {
	accountUserApp application.AccountUserUseCase
}

func NewAccountUserHandler(accountUserApp application.AccountUserUseCase) *AccountUserHandler {
	return &AccountUserHandler{
		accountUserApp: accountUserApp,
	}
}

func (obj *AccountUserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("search users request")

	_, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountIdStr := r.URL.Query().Get("accountId")
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	query := strings.TrimSpace(secure.SanitizeXss(r.URL.Query().Get("query")))
	if query == "" {
		response := web_helpers.NewBadRequestErrorResponse("Запрос не может быть пустым")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	users, err := obj.accountUserApp.SearchUsers(r.Context(), accountId, query, 5)
	if err != nil {
		log.Error("failed to search users", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewSearchUsersResponse(users)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) Invite(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("invite user request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body models.InviteRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Неверный формат запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	query := strings.TrimSpace(secure.SanitizeXss(body.Query))
	if query == "" {
		response := web_helpers.NewBadRequestErrorResponse("Запрос не может быть пустым")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	users, err := obj.accountUserApp.SearchUsers(r.Context(), accountId, query, 1)
	if err != nil {
		log.Error("failed to search user", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if len(users) == 0 {
		response := web_helpers.NewNotFoundErrorResponse("Пользователь не найден")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	invite, err := obj.accountUserApp.CreateInvite(r.Context(), accountId, authUser.Id, users[0].Id)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrSelfInvite):
			response := web_helpers.NewBadRequestErrorResponse("Нельзя пригласить самого себя")
			web_helpers.WriteResponseJSON(w, response.Code, response)
		case errors.Is(err, application.ErrInviteAlreadyExists):
			response := web_helpers.NewBadRequestErrorResponse("Приглашение уже отправлено")
			web_helpers.WriteResponseJSON(w, response.Code, response)
		case errors.Is(err, application.ErrAlreadyMember):
			response := web_helpers.NewBadRequestErrorResponse("Пользователь уже является участником счёта")
			web_helpers.WriteResponseJSON(w, response.Code, response)
		case errors.Is(err, application.ErrNotOwner):
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
		default:
			log.Error("failed to create invite", zap.Error(err))
			response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
			web_helpers.WriteResponseJSON(w, response.Code, response)
		}
		return
	}

	response := web_helpers.NewInviteResponse(invite)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get members request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	members, err := obj.accountUserApp.GetMembers(r.Context(), accountId, authUser.Id)
	if err != nil {
		log.Error("failed to get members", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewMembersResponse(members)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("accept invite request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	invite, err := obj.accountUserApp.AcceptInvite(r.Context(), accountId, authUser.Id)
	if err != nil {
		if errors.Is(err, application.ErrInviteNotFound) {
			response := web_helpers.NewNotFoundErrorResponse("Приглашение не найдено")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		log.Error("failed to accept invite", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewInviteResponse(invite)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) RejectInvite(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("reject invite request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err = obj.accountUserApp.RejectInvite(r.Context(), accountId, authUser.Id); err != nil {
		if errors.Is(err, application.ErrInviteNotFound) {
			response := web_helpers.NewNotFoundErrorResponse("Приглашение не найдено")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		log.Error("failed to reject invite", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("remove member request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	userId, err := userIDFromMembersPath(r)
	if err != nil || userId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id пользователя")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err = obj.accountUserApp.RemoveMember(r.Context(), accountId, authUser.Id, userId); err != nil {
		switch {
		case errors.Is(err, application.ErrNotOwner):
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
		case errors.Is(err, application.ErrCannotRemoveOwner):
			response := web_helpers.NewBadRequestErrorResponse("Невозможно удалить владельца счёта")
			web_helpers.WriteResponseJSON(w, response.Code, response)
		default:
			log.Error("failed to remove member", zap.Error(err))
			response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
			web_helpers.WriteResponseJSON(w, response.Code, response)
		}
		return
	}

	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountUserHandler) GetPendingInvites(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get pending invites request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	invites, err := obj.accountUserApp.GetPendingInvites(r.Context(), authUser.Id)
	if err != nil {
		log.Error("failed to get pending invites", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewPendingInvitesResponse(invites)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

// POST /api/accounts/{accountId}/leave
func (obj *AccountUserHandler) LeaveAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("leave account request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := accountIDFromPath(r)
	if err != nil || accountId < 1 {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err = obj.accountUserApp.LeaveAccount(r.Context(), accountId, authUser.Id); err != nil {
		if errors.Is(err, application.ErrOwnerCannotLeave) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		log.Error("failed to leave account", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
