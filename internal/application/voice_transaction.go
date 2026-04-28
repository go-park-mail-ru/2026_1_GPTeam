package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

//go:generate mockgen -source=voice_transaction.go -destination=mocks/voice_transaction.go -package=mocks
type AIConsultantClient interface {
	Transcribe(ctx context.Context, audioData []byte, filename string) (string, error)
	ParseTransaction(ctx context.Context, transcript string, types, categories []string) (*models.TransactionDraft, error)
}

type VoiceTransactionUseCase interface {
	CreateVoiceTransaction(ctx context.Context, audioData []byte, filename string) (*models.TransactionDraft, error)
}

type VoiceTransactionService struct {
	client AIConsultantClient
	enums  EnumsUseCase
}

func NewVoiceTransactionService(client AIConsultantClient, enums EnumsUseCase) *VoiceTransactionService {
	return &VoiceTransactionService{
		client: client,
		enums:  enums,
	}
}

func (s *VoiceTransactionService) CreateVoiceTransaction(ctx context.Context, audioData []byte, filename string) (*models.TransactionDraft, error) {
	transcript, err := s.client.Transcribe(ctx, audioData, filename)
	if err != nil {
		return nil, err
	}

	draft, err := s.client.ParseTransaction(
		ctx,
		transcript,
		s.enums.GetTransactionTypes(),
		s.enums.GetCategoryTypes(),
	)
	if err != nil {
		return nil, err
	}

	return draft, nil
}
