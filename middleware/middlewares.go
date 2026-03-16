package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	base2 "github.com/go-park-mail-ru/2026_1_GPTeam/web/base"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base2.SetCORS(w)
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler, authUseCase auth.AuthenticationServiceInterface, userUseCase application.UserUseCaseInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/auth/") && path != "/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}

		isAuth, userID := authUseCase.IsAuth(r)
		if !isAuth {
			fmt.Printf("[Auth] 401 для пути: %s\n", path)
			response := base2.NewUnauthorizedErrorResponse()
			base2.WriteResponseJSON(w, response.Code, response)
			return
		}

		authUser, err := userUseCase.GetById(context.Background(), userID)
		if err != nil {
			response := base2.NewUnauthorizedErrorResponse()
			base2.WriteResponseJSON(w, response.Code, response)
			return
		}

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
			response := base2.NewMethodError()
			base2.WriteResponseJSON(w, response.Code, response)
		})
	}
}
