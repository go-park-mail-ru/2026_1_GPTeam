package web

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

const maxAudioSize = 25 << 20

type VoiceHandler struct {
	voiceSvc application.VoiceTransactionUseCase
	enumsApp application.EnumsUseCase
}

func NewVoiceHandler(voiceSvc application.VoiceTransactionUseCase, es application.EnumsUseCase) *VoiceHandler {
	return &VoiceHandler{
		voiceSvc: voiceSvc,
		enumsApp: es,
	}
}

func (h *VoiceHandler) CreateVoiceTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("voice: create transaction request")

	_, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("voice: unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err := r.ParseMultipartForm(maxAudioSize); err != nil {
		log.Warn("voice: multipart error", zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("audio", "Файл слишком большой или форма некорректна"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		log.Warn("voice: file missing", zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("audio", "Поле 'audio' обязательно"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(io.LimitReader(file, maxAudioSize))
	if err != nil {
		log.Error("voice: read audio error", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(ctx))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	recordedAt := time.Now()
	if raw := r.FormValue("recorded_at"); raw != "" {
		if t, parseErr := time.Parse(time.RFC3339, raw); parseErr == nil {
			recordedAt = t
		}
	}

	groqCtx, groqCancel := context.WithTimeout(context.WithoutCancel(ctx), 90*time.Second)
	defer groqCancel()

	draft, err := h.voiceSvc.CreateVoiceTransaction(groqCtx, audioData, header.Filename)
	if err != nil {
		log.Error("voice: processing failed", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(ctx))
		response.Message = "Не удалось обработать аудио"
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if draft == nil || (draft.Title == "" && draft.Value == 0) {
		log.Warn("voice: no transaction in text")
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		response.Code = http.StatusUnprocessableEntity
		response.Message = "В вашей речи не найдено данных о транзакции"
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	valReq := web_helpers.TransactionRequest{
		Title:           draft.Title,
		Description:     draft.Description,
		Value:           draft.Value,
		Type:            draft.Type,
		Category:        draft.Category,
		TransactionDate: draft.Date,
	}

	validationErrors := validators.ValidateTransaction(
		valReq,
		h.enumsApp.GetTransactionTypes(),
		h.enumsApp.GetCategoryTypes(),
		h.enumsApp.GetCurrencyCodes(),
	)

	if len(validationErrors) > 0 {
		log.Warn("voice: validation failed for draft", zap.Any("errors", validationErrors))
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	draft.RecordedAt = recordedAt
	log.Info("voice: draft created successfully")

	resp := web_helpers.NewVoiceTransactionDraftResponse(web_helpers.TransactionDraftData{
		RawText:     draft.RawText,
		Value:       draft.Value,
		Type:        draft.Type,
		Category:    draft.Category,
		Currency:    draft.Currency,
		Title:       draft.Title,
		Description: draft.Description,
		RecordedAt:  draft.RecordedAt,
		Date:        draft.Date,
	})
	web_helpers.WriteResponseJSON(w, http.StatusOK, resp)
}
