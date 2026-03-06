package middleware

import (
	"context"
	"main/auth"
	"main/base"
	"main/storage"
	"net/http"
	"strconv"
	"strings"
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
		if strings.HasPrefix(r.URL.Path, "/auth/") && r.URL.Path != "/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}
		isAuth, userID := auth.IsAuth(r)
		if !isAuth {
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
