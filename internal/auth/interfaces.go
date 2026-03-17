package auth

import (
	"context"
	"net/http"
)

type AuthenticationServiceInterface interface {
	GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userId int)
	IsAuth(r *http.Request) (bool, int)
	ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int)
}
