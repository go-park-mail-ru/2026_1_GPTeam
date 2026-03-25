package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web_helpers.SetCORS(w)
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler, authService auth.AuthenticationService, userApp application.UserUseCase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		path := r.URL.Path
		log.Info("[auth middleware] checking", zap.String("path", path))
		if (strings.HasPrefix(path, "/auth/") && path != "/auth/logout") || strings.HasPrefix(path, "/enums/") {
			log.Info("[auth middleware] pass without checking", zap.String("path", path))
			next.ServeHTTP(w, r)
			return
		}

		isAuth, userId := authService.IsAuth(r)
		if !isAuth {
			log.Warn("[auth middleware] auth check failed", zap.String("path", path))
			response := web_helpers.NewUnauthorizedErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		log.Info("[auth middleware] auth service success check", zap.String("path", path), zap.Int("user_id", userId))

		authUser, err := userApp.GetById(r.Context(), userId)
		if err != nil {
			log.Error("[auth middleware] get user by id from db failed", zap.String("path", path), zap.Int("user_id", userId), zap.Error(err))
			response := web_helpers.NewUnauthorizedErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		log.Info("[auth middleware] auth success", zap.String("path", path), zap.Int("user_id", userId))

		ctx := context.WithValue(r.Context(), "user", authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MethodValidationMiddleware(allowedMethods ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(allowedMethods, r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			response := web_helpers.NewMethodError()
			web_helpers.WriteResponseJSON(w, response.Code, response)
		})
	}
}

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		defer func() {
			if err := recover(); err != nil {
				log.Error("[panic middleware] panic", zap.Any("err", err))
				response := web_helpers.NewServerErrorResponse("")
				web_helpers.WriteResponseJSON(w, response.Code, response)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
