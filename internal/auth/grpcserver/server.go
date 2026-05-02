package grpcserver

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC AuthService поверх JwtUseCase.
type Server struct {
	authv1.UnimplementedAuthServiceServer
	JWT jwt_auth.JwtUseCase
}

func deviceOrDefault(deviceID string) string {
	if deviceID == "" {
		return "pass"
	}
	return deviceID
}

func (s *Server) IssueTokens(ctx context.Context, req *authv1.IssueTokensRequest) (*authv1.IssueTokensResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be positive")
	}
	uid := int(req.GetUserId())
	access, err := s.JWT.GenerateToken(uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate access token: %v", err)
	}
	refresh, err := s.JWT.GenerateRefreshToken(ctx, uid, deviceOrDefault(req.GetDeviceId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate refresh token: %v", err)
	}
	return &authv1.IssueTokensResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Server) ValidateAccess(_ context.Context, req *authv1.ValidateAccessRequest) (*authv1.ValidateAccessResponse, error) {
	if req.GetAccessToken() == "" {
		return &authv1.ValidateAccessResponse{Valid: false}, nil
	}
	ok, uid := s.JWT.CheckToken(req.GetAccessToken())
	if !ok {
		return &authv1.ValidateAccessResponse{Valid: false}, nil
	}
	return &authv1.ValidateAccessResponse{Valid: true, UserId: int32(uid)}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return &authv1.RefreshResponse{Valid: false}, nil
	}
	ok, userID := s.JWT.CheckRefreshToken(ctx, req.GetRefreshToken())
	if !ok {
		return &authv1.RefreshResponse{Valid: false}, nil
	}
	s.JWT.DeleteRefreshToken(ctx, req.GetRefreshToken())
	access, err := s.JWT.GenerateToken(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate access token: %v", err)
	}
	refresh, err := s.JWT.GenerateRefreshToken(ctx, userID, deviceOrDefault(req.GetDeviceId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate refresh token: %v", err)
	}
	return &authv1.RefreshResponse{
		Valid:        true,
		UserId:       int32(userID),
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Server) Revoke(ctx context.Context, req *authv1.RevokeRequest) (*authv1.RevokeResponse, error) {
	if req.GetRefreshToken() != "" {
		s.JWT.DeleteRefreshToken(ctx, req.GetRefreshToken())
	}
	return &authv1.RevokeResponse{}, nil
}
