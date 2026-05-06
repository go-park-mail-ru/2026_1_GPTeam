package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"go.uber.org/zap"
)

const defaultDeviceID = "pass"

// GrpcAuthAdapter: выдача/refresh/revoke через gRPC, проверка access JWT локально (без RPC на каждый запрос).
type GrpcAuthAdapter struct {
	client  authv1.AuthServiceClient
	secret  []byte
	version string
}

func NewGrpcAuthAdapter(client authv1.AuthServiceClient, jwtSecret, jwtVersion string) *GrpcAuthAdapter {
	return &GrpcAuthAdapter{
		client:  client,
		secret:  []byte(jwtSecret),
		version: jwtVersion,
	}
}

func (a *GrpcAuthAdapter) GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userID int) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("generating new auth for user (grpc)",
		zap.Int("user_id", userID))
	t0 := time.Now()
	resp, err := a.client.IssueTokens(ctx, &authv1.IssueTokensRequest{
		UserId:   int32(userID),
		DeviceId: defaultDeviceID,
	})
	if err != nil {
		log.Error("IssueTokens failed",
			zap.Error(err),
			zap.Duration("grpc_duration", time.Since(t0)))
		return
	}
	WriteAuthCookies(w, resp.GetAccessToken(), resp.GetRefreshToken())
	log.Info("grpc auth cookies set",
		zap.Int("user_id", userID),
		zap.Duration("grpc_duration", time.Since(t0)))
}

func (a *GrpcAuthAdapter) IsAuth(ctx context.Context, r *http.Request) (bool, int) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("checking if user authenticated (local jwt)")
	cookie, err := r.Cookie(TokenName)
	if err != nil {
		return false, -1
	}
	isValid, userId := jwt_auth.ValidateAccessToken(cookie.Value, a.secret, a.version)
	appMetrics := metrics.GetMetrics()
	label := "ok"
	if !isValid {
		label = "fail"
	}
	appMetrics.AuthValidateTokenTotal.WithLabelValues(label).Inc()
	return isValid, userId
}

func (a *GrpcAuthAdapter) ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("clear old token cookie (grpc revoke)")
	if cookie, err := r.Cookie(RefreshTokenName); err == nil && cookie.Value != "" {
		if _, err := a.client.Revoke(ctx, &authv1.RevokeRequest{RefreshToken: cookie.Value}); err != nil {
			log.Warn("Revoke failed",
				zap.Error(err))
		}
	}
	ClearAuthCookies(w)
	log.Info("auth cookies cleared")
}

func (a *GrpcAuthAdapter) Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int) {
	log := logger.GetLoggerWithRequestId(ctx)
	log.Info("refresh token (grpc)")
	cookie, err := r.Cookie(RefreshTokenName)
	if err != nil {
		return false, -1
	}
	t0 := time.Now()
	resp, err := a.client.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: cookie.Value,
		DeviceId:     defaultDeviceID,
	})
	if err != nil {
		log.Warn("Refresh grpc failed",
			zap.Error(err),
			zap.Duration("grpc_duration", time.Since(t0)))
		return false, -1
	}
	log.Info("Refresh grpc ok",
		zap.Duration("grpc_duration", time.Since(t0)))
	appMetrics := metrics.GetMetrics()
	label := "ok"
	if !resp.GetValid() {
		label = "fail"
	}
	appMetrics.AuthValidateRefreshTokenTotal.WithLabelValues(label).Inc()
	if !resp.GetValid() {
		return false, -1
	}
	WriteAuthCookies(w, resp.GetAccessToken(), resp.GetRefreshToken())
	return true, int(resp.GetUserId())
}
