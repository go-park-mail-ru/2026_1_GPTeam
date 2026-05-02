package ai

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai/groq"
)

type AiService interface {
	Transcribe(ctx context.Context, audioData []byte, filename string) (string, error)
	ParseTransaction(ctx context.Context, transcript string, types, categories []string) (*groq.TransactionDraft, error)
}

type GroqAiService struct {
	client *groq.GroqClient
}

func NewGroqAiService(client *groq.GroqClient) *GroqAiService {
	return &GroqAiService{
		client: client,
	}
}

func (s *GroqAiService) Transcribe(ctx context.Context, audioData []byte, filename string) (string, error) {
	return s.client.Transcribe(ctx, audioData, filename)
}

func (s *GroqAiService) ParseTransaction(ctx context.Context, transcript string, types, categories []string) (*groq.TransactionDraft, error) {
	return s.client.ParseTransaction(ctx, transcript, types, categories)
}
