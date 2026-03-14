package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base.SetCORS(w)
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		isPublicPath := path == "/" ||
			path == "/login" ||
			path == "/signup" ||
			(strings.HasPrefix(path, "/auth/") && path != "/auth/logout")

		if isPublicPath {
			next.ServeHTTP(w, r)
			return
		}

		isAuth, userID := auth.IsAuth(r)
		if !isAuth {
			fmt.Printf("[Auth] 401 для пути: %s\n", path)
			response := base.NewUnauthorizedErrorResponse()
			base.WriteResponseJSON(w, response.Code, response)
			return
		}

		id, err := strconv.Atoi(userID)
		if err != nil {
			response := base.NewUnauthorizedErrorResponse()
			base.WriteResponseJSON(w, response.Code, response)
			return
		}

		authUser, exists := storage.GetUserByID(id)
		if !exists {
			response := base.NewUnauthorizedErrorResponse()
			base.WriteResponseJSON(w, response.Code, response)
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
			response := base.NewMethodError()
			base.WriteResponseJSON(w, response.Code, response)
		})
	}
}
