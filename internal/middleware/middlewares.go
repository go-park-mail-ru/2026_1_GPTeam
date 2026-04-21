package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure/rate_limiter"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (obj *responseWriter) WriteHeader(code int) {
	obj.StatusCode = code
	obj.ResponseWriter.WriteHeader(code)
}

func (obj *responseWriter) Write(b []byte) (int, error) {
	if obj.StatusCode == 0 {
		obj.StatusCode = http.StatusOK
	}
	return obj.ResponseWriter.Write(b)
}

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
		log := logger.GetLoggerWithRequestId(r.Context())
		path := r.URL.Path
		log.Info("[auth middleware] checking",
			zap.String("path", path))
		if (strings.HasPrefix(path, "/auth/") && path != "/auth/logout") || strings.HasPrefix(path, "/enums/") || strings.HasPrefix(path, "/img/") {
			log.Info("[auth middleware] pass without checking",
				zap.String("path", path))
			next.ServeHTTP(w, r)
			return
		}

		isAuth, userId := authService.IsAuth(r.Context(), r)
		if !isAuth {
			log.Warn("[auth middleware] auth check failed",
				zap.String("path", path))
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
		log.Info("[auth middleware] auth success",
			zap.String("path", path),
			zap.Int("user_id", userId))

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
				requestId, ok := r.Context().Value("request_id").(string)
				if !ok {
					log.Error("[panic middleware] panic",
						zap.Any("err", err))
				} else {
					log.Error("[panic middleware] panic",
						zap.String("request_id", requestId),
						zap.Any("err", err))
				}
				response := web_helpers.NewServerErrorResponse(requestId)
				web_helpers.WriteResponseJSON(w, response.Code, response)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetAccessLogger()
		requestId := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestId)
		log.Info("Request",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", requestId))
		wr := &responseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		timeStart := time.Now()
		next.ServeHTTP(wr, r.WithContext(ctx))
		duration := time.Since(timeStart)
		log.Info("Response",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("request_id", requestId),
			zap.Int("status_code", wr.StatusCode),
			zap.String("duration", duration.String()))
	})
}

func CSRFMiddleware(next http.Handler, csrfService secure.CsrfService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			next.ServeHTTP(w, r)
			return
		}
		log := logger.GetLoggerWithRequestId(r.Context())
		if !csrfService.ValidateSecFetchSite(r) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch {
			accessToken, err := csrfService.GetAccessToken(r.Context(), r)
			if err != nil {
				log.Warn("[CSRF middleware] no access token in cookie")
				response := web_helpers.NewForbiddenErrorResponse()
				web_helpers.WriteResponseJSON(w, response.Code, response)
				return
			}
			csrfCookieToken := csrfService.GetCsrfFromCookie(r.Context(), r)
			if csrfCookieToken == "" {
				log.Warn("[CSRF middleware] no CSRF cookie in request")
				response := web_helpers.NewForbiddenErrorResponse()
				web_helpers.WriteResponseJSON(w, response.Code, response)
				return
			}
			isValid, err := csrfService.ValidateCsrf(r.Context(), csrfCookieToken, accessToken)
			if err != nil || !isValid {
				log.Warn("[CSRF middleware] invalid CSRF cookie in request")
				response := web_helpers.NewForbiddenErrorResponse()
				web_helpers.WriteResponseJSON(w, response.Code, response)
				return
			}
			csrfHeaderToken := csrfService.GetCsrfFromHeader(r)
			if csrfHeaderToken == "" {
				log.Warn("[CSRF middleware] no CSRF header in request")
				response := web_helpers.NewForbiddenErrorResponse()
				web_helpers.WriteResponseJSON(w, response.Code, response)
				return
			}
			if csrfHeaderToken != csrfCookieToken {
				log.Warn("[CSRF middleware] invalid CSRF header in request")
				response := web_helpers.NewForbiddenErrorResponse()
				web_helpers.WriteResponseJSON(w, response.Code, response)
				return
			}
		}
		csrfService.SetCsrfCookie(r.Context(), w, r)
		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(next http.Handler, rateLimiter rate_limiter.RateLimiterInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLoggerWithRequestId(r.Context())
		ip, err := rate_limiter.GetRealIp(r)
		if err != nil {
			log.Warn("[rate limit middleware] unable to get ip - return")
			response := web_helpers.NewTooManyRequestsResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if rateLimiter.IsTrustedIp(ip) {
			log.Info("[rate limit middleware] trusted ip - skip",
				zap.String("ip", ip))
			next.ServeHTTP(w, r)
			return
		}
		isBlocked := rateLimiter.IsIpBlocked(r.Context(), ip)
		if isBlocked {
			log.Warn("[rate limit middleware] ip blocked",
				zap.String("ip", ip))
			response := web_helpers.NewTooManyRequestsResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		isAllowed := rateLimiter.Allow(r.Context(), ip)
		if !isAllowed {
			log.Warn("[rate limit middleware] ip exceeded limit",
				zap.String("ip", ip))
			rateLimiter.BlockIp(r.Context(), ip)
			response := web_helpers.NewTooManyRequestsResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
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
