package ai

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/groq"
	aiv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type GrpcAiAdapter struct {
	client aiv1.AiServiceClient
}

func NewGrpcAiAdapter(client aiv1.AiServiceClient) *GrpcAiAdapter {
	return &GrpcAiAdapter{
		client: client,
	}
}

func (a *GrpcAiAdapter) Transcribe(ctx context.Context, audioData []byte, filename string) (string, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("transcribing audio via gRPC")

	resp, err := a.client.Transcribe(ctx, &aiv1.TranscribeRequest{
		AudioData: audioData,
		Filename:  filename,
	})
	if err != nil {
		log.Error("transcribe gRPC failed", zap.Error(err))
		return "", err
	}

	log.Info("transcribe gRPC success")
	return resp.GetText(), nil
}

func (a *GrpcAiAdapter) ParseTransaction(ctx context.Context, transcript string, types, categories []string) (*groq.TransactionDraft, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("parsing transaction via gRPC")

	resp, err := a.client.ParseTransaction(ctx, &aiv1.ParseTransactionRequest{
		Transcript: transcript,
		Types:      types,
		Categories: categories,
	})
	if err != nil {
		log.Error("parse transaction gRPC failed", zap.Error(err))
		return nil, err
	}

	draft := resp.GetDraft()
	if draft == nil {
		return nil, nil
	}

	return &groq.TransactionDraft{
		RawText:     transcript,
		Value:       draft.GetValue(),
		Type:        draft.GetType(),
		Category:    draft.GetCategory(),
		Title:       draft.GetTitle(),
		Description: draft.GetDescription(),
		Date:        parseDate(draft.GetDate()),
	}, nil
}

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}
