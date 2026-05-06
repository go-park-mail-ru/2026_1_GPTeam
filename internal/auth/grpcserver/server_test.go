package grpcserver

import (
	"context"
	"errors"
	"testing"

	jwtmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth/mocks"
	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeviceOrDefault(t *testing.T) {
	t.Parallel()
	require.Equal(t, "pass", deviceOrDefault(""))
	require.Equal(t, "iphone", deviceOrDefault("iphone"))
}

func TestServer_IssueTokens(t *testing.T) {
	t.Parallel()

	t.Run("invalid user_id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		s := &Server{JWT: jwt}

		resp, err := s.IssueTokens(context.Background(), &authv1.IssueTokensRequest{UserId: 0})

		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("generate access fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().GenerateToken(7).Return("", errors.New("err"))
		s := &Server{JWT: jwt}

		resp, err := s.IssueTokens(context.Background(), &authv1.IssueTokensRequest{UserId: 7})

		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("generate refresh fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().GenerateToken(7).Return("at", nil)
		jwt.EXPECT().GenerateRefreshToken(gomock.Any(), 7, "pass").Return("", errors.New("err"))
		s := &Server{JWT: jwt}

		resp, err := s.IssueTokens(context.Background(), &authv1.IssueTokensRequest{UserId: 7})

		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("ok with custom device", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().GenerateToken(7).Return("at", nil)
		jwt.EXPECT().GenerateRefreshToken(gomock.Any(), 7, "ios").Return("rt", nil)
		s := &Server{JWT: jwt}

		resp, err := s.IssueTokens(context.Background(), &authv1.IssueTokensRequest{UserId: 7, DeviceId: "ios"})

		require.NoError(t, err)
		require.Equal(t, "at", resp.GetAccessToken())
		require.Equal(t, "rt", resp.GetRefreshToken())
	})
}

func TestServer_ValidateAccess(t *testing.T) {
	t.Parallel()

	t.Run("empty token -> invalid, no error", func(t *testing.T) {
		s := &Server{}
		resp, err := s.ValidateAccess(context.Background(), &authv1.ValidateAccessRequest{})
		require.NoError(t, err)
		require.False(t, resp.GetValid())
	})

	t.Run("invalid token -> invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckToken("tok").Return(false, -1)
		s := &Server{JWT: jwt}

		resp, err := s.ValidateAccess(context.Background(), &authv1.ValidateAccessRequest{AccessToken: "tok"})
		require.NoError(t, err)
		require.False(t, resp.GetValid())
	})

	t.Run("valid token -> valid + user_id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckToken("tok").Return(true, 42)
		s := &Server{JWT: jwt}

		resp, err := s.ValidateAccess(context.Background(), &authv1.ValidateAccessRequest{AccessToken: "tok"})
		require.NoError(t, err)
		require.True(t, resp.GetValid())
		require.Equal(t, int32(42), resp.GetUserId())
	})
}

func TestServer_Refresh(t *testing.T) {
	t.Parallel()

	t.Run("empty refresh token -> invalid", func(t *testing.T) {
		s := &Server{}
		resp, err := s.Refresh(context.Background(), &authv1.RefreshRequest{})
		require.NoError(t, err)
		require.False(t, resp.GetValid())
	})

	t.Run("invalid refresh -> invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckRefreshToken(gomock.Any(), "rt").Return(false, -1)
		s := &Server{JWT: jwt}

		resp, err := s.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "rt"})
		require.NoError(t, err)
		require.False(t, resp.GetValid())
	})

	t.Run("valid + access generation fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckRefreshToken(gomock.Any(), "rt").Return(true, 5)
		jwt.EXPECT().DeleteRefreshToken(gomock.Any(), "rt")
		jwt.EXPECT().GenerateToken(5).Return("", errors.New("err"))
		s := &Server{JWT: jwt}

		resp, err := s.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "rt"})
		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("valid + refresh generation fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckRefreshToken(gomock.Any(), "rt").Return(true, 5)
		jwt.EXPECT().DeleteRefreshToken(gomock.Any(), "rt")
		jwt.EXPECT().GenerateToken(5).Return("at", nil)
		jwt.EXPECT().GenerateRefreshToken(gomock.Any(), 5, "pass").Return("", errors.New("err"))
		s := &Server{JWT: jwt}

		resp, err := s.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "rt"})
		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().CheckRefreshToken(gomock.Any(), "rt").Return(true, 5)
		jwt.EXPECT().DeleteRefreshToken(gomock.Any(), "rt")
		jwt.EXPECT().GenerateToken(5).Return("new_at", nil)
		jwt.EXPECT().GenerateRefreshToken(gomock.Any(), 5, "android").Return("new_rt", nil)
		s := &Server{JWT: jwt}

		resp, err := s.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "rt", DeviceId: "android"})
		require.NoError(t, err)
		require.True(t, resp.GetValid())
		require.Equal(t, int32(5), resp.GetUserId())
		require.Equal(t, "new_at", resp.GetAccessToken())
		require.Equal(t, "new_rt", resp.GetRefreshToken())
	})
}

func TestServer_Revoke(t *testing.T) {
	t.Parallel()

	t.Run("empty token -> nothing called", func(t *testing.T) {
		s := &Server{}
		resp, err := s.Revoke(context.Background(), &authv1.RevokeRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("with token -> DeleteRefreshToken called", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		jwt := jwtmocks.NewMockJwtUseCase(ctrl)
		jwt.EXPECT().DeleteRefreshToken(gomock.Any(), "rt")
		s := &Server{JWT: jwt}

		resp, err := s.Revoke(context.Background(), &authv1.RevokeRequest{RefreshToken: "rt"})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}
