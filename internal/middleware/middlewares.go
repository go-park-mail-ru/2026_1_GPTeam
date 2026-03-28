package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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

func NoDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path == "/img/" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler, authService auth.AuthenticationService, userApp application.UserUseCase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if (strings.HasPrefix(path, "/auth/") && path != "/auth/logout") || strings.HasPrefix(path, "/enums/") || strings.HasPrefix(path, "/img/") {
			next.ServeHTTP(w, r)
			return
		}

		isAuth, userId := authService.IsAuth(r)
		if !isAuth {
			fmt.Printf("[Auth] 401 для пути: %s\n", path)
			response := web_helpers.NewUnauthorizedErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}

		authUser, err := userApp.GetById(r.Context(), userId)
		if err != nil {
			response := web_helpers.NewUnauthorizedErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
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
			response := web_helpers.NewMethodError()
			web_helpers.WriteResponseJSON(w, response.Code, response)
		})
	}
}

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[Panic] %s\n", err)
				response := web_helpers.NewServerErrorResponse("")
				web_helpers.WriteResponseJSON(w, response.Code, response)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// мидлварки для тестов
func SetUserCtx(ctx context.Context, user models.UserModel) context.Context {
	return context.WithValue(ctx, models.UserContextKey, user)
}

func GetUserCtx(ctx context.Context) (models.UserModel, bool) {
	user, ok := ctx.Value(models.UserContextKey).(models.UserModel)
	return user, ok
}
