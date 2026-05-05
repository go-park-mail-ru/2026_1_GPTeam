package grpcserver

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/application"
	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=server.go -destination=mocks/mock_server.go -package=mocks
type AvatarUseCase interface {
	Upload(ctx context.Context, data []byte, extension string) (string, error)
}

type Server struct {
	fsv1.UnimplementedFileServiceServer
	avatars AvatarUseCase
}

func NewServer(avatars AvatarUseCase) *Server {
	return &Server{avatars: avatars}
}

func (s *Server) Upload(ctx context.Context, req *fsv1.UploadRequest) (*fsv1.UploadResponse, error) {
	data := req.GetData()
	if len(data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	startTime := time.Now()
	name, err := s.avatars.Upload(ctx, data, req.GetExtension())
	duration := time.Since(startTime)
	appMetrics := metrics.GetMetrics()
	label := "ok"
	if err != nil {
		label = "error"
	}
	appMetrics.FsAvatarUploadDuration.WithLabelValues(label).Observe(float64(duration.Milliseconds()))
	switch {
	case err == nil:
		return &fsv1.UploadResponse{Filename: name}, nil
	case errors.Is(err, application.ErrEmptyData),
		errors.Is(err, application.ErrTooLarge):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, status.Errorf(codes.Internal, "upload failed: %v", err)
	}
}
