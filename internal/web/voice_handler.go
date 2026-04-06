package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const maxAudioSize = 25 << 20 // 25 MB — лимит Groq API

// VoiceHandler обрабатывает голосовой ввод транзакций.
// Проксирует аудио в Groq Whisper (STT), затем парсит текст через Groq LLaMA.
type VoiceHandler struct {
	transcription *application.TranscriptionService
	parser        *application.ParserService
}

// NewVoiceHandler создаёт хэндлер голосового ввода.
func NewVoiceHandler(ts *application.TranscriptionService, ps *application.ParserService) *VoiceHandler {
	return &VoiceHandler{
		transcription: ts,
		parser:        ps,
	}
}

// CreateVoiceTransaction принимает аудио-файл (multipart/form-data, поле "audio"),
// транскрибирует через Groq Whisper и парсит в черновик транзакции через Groq LLaMA.
// Возвращает VoiceTransactionDraftResponse для предзаполнения формы на клиенте.
// Сохранение в БД — через стандартный POST /transactions после подтверждения пользователем.
func (h *VoiceHandler) CreateVoiceTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerWIthRequestId(ctx)

	if err := r.ParseMultipartForm(maxAudioSize); err != nil {
		log.Warn("voice: failed to parse multipart form", zap.Error(err))
		writeVoiceJSON(w, http.StatusBadRequest,
			web_helpers.NewVoiceErrorResponse(http.StatusBadRequest, "Файл слишком большой или форма некорректна"))
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		log.Warn("voice: field 'audio' missing", zap.Error(err))
		writeVoiceJSON(w, http.StatusBadRequest,
			web_helpers.NewVoiceErrorResponse(http.StatusBadRequest, "Поле 'audio' обязательно"))
		return
	}
	defer file.Close()

	if header.Size > maxAudioSize {
		writeVoiceJSON(w, http.StatusRequestEntityTooLarge,
			web_helpers.NewVoiceErrorResponse(http.StatusRequestEntityTooLarge, "Аудио превышает 25 МБ"))
		return
	}

	audioData, err := io.ReadAll(io.LimitReader(file, maxAudioSize))
	if err != nil {
		log.Error("voice: failed to read audio data", zap.Error(err))
		writeVoiceJSON(w, http.StatusInternalServerError,
			web_helpers.NewVoiceErrorResponse(http.StatusInternalServerError, "Ошибка чтения аудио"))
		return
	}

	recordedAt := time.Now()
	if raw := r.FormValue("recorded_at"); raw != "" {
		if t, parseErr := time.Parse(time.RFC3339, raw); parseErr == nil {
			recordedAt = t
		}
	}

	// Отвязываем от контекста запроса — браузер может закрыть соединение
	// пока Groq обрабатывает аудио (STT занимает 5-15 сек).
	// context.WithoutCancel сохраняет values (request_id) но игнорирует отмену.
	groqCtx, groqCancel := context.WithTimeout(context.WithoutCancel(ctx), 90*time.Second)
	defer groqCancel()

	transcript, err := h.transcription.Transcribe(groqCtx, audioData, header.Filename)
	if err != nil {
		log.Error("voice: transcription failed",
			zap.String("filename", header.Filename),
			zap.Error(err))
		writeVoiceJSON(w, http.StatusInternalServerError,
			web_helpers.NewVoiceErrorResponse(http.StatusInternalServerError, "Ошибка распознавания речи: "+err.Error()))
		return
	}

	draft, err := h.parser.ParseTransaction(groqCtx, transcript)
	if err != nil {
		log.Error("voice: parsing failed",
			zap.String("transcript", transcript),
			zap.Error(err))
		writeVoiceJSON(w, http.StatusUnprocessableEntity,
			web_helpers.NewVoiceErrorResponse(http.StatusUnprocessableEntity,
				"Не удалось извлечь данные транзакции: "+err.Error()))
		return
	}
	draft.RecordedAt = recordedAt

	resp := web_helpers.NewVoiceTransactionDraftResponse(web_helpers.TransactionDraftData{
		RawText:     draft.RawText,
		Value:       draft.Value,
		Type:        draft.Type,
		Category:    draft.Category,
		Currency:    draft.Currency,
		Title:       draft.Title,
		Description: draft.Description,
		RecordedAt:  draft.RecordedAt,
	})
	writeVoiceJSON(w, http.StatusOK, resp)
}

func writeVoiceJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
