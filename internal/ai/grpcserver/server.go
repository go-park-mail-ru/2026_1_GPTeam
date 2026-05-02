package grpcserver

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/ai"
	aiv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/ai/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	aiv1.UnimplementedAiServiceServer
	AI ai.AiService
}

func (s *Server) Transcribe(ctx context.Context, req *aiv1.TranscribeRequest) (*aiv1.TranscribeResponse, error) {
	if len(req.GetAudioData()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "audio_data is required")
	}

	text, err := s.AI.Transcribe(ctx, req.GetAudioData(), req.GetFilename())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transcribe failed: %v", err)
	}

	return &aiv1.TranscribeResponse{
		Text: text,
	}, nil
}

func (s *Server) ParseTransaction(ctx context.Context, req *aiv1.ParseTransactionRequest) (*aiv1.ParseTransactionResponse, error) {
	if req.GetTranscript() == "" {
		return nil, status.Error(codes.InvalidArgument, "transcript is required")
	}

	draft, err := s.AI.ParseTransaction(ctx, req.GetTranscript(), req.GetTypes(), req.GetCategories())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "parse transaction failed: %v", err)
	}

	if draft == nil {
		return nil, status.Error(codes.NotFound, "transaction draft not found")
	}

	return &aiv1.ParseTransactionResponse{
		Draft: &aiv1.TransactionDraft{
			Value:       draft.Value,
			Type:        draft.Type,
			Category:    draft.Category,
			Title:       draft.Title,
			Description: draft.Description,
			Date:        draft.Date.Format("2006-01-02"),
		},
	}, nil
}
