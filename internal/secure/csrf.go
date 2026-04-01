package secure

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const CsrfCookieName = "csrf_token"
const CsrfHeaderName = "X-CSRF-Token"
const CsrfCookieExpirationTime = time.Hour

type CsrfService interface {
	GenerateRandValue(ctx context.Context) (string, error)
	GenerateCsrf(ctx context.Context, token string) (string, error)
	ValidateCsrf(ctx context.Context, csrf string, token string) (bool, error)
	GetCsrfValueFromToken(ctx context.Context, csrf string) (string, error)
	SetCsrfCookie(ctx context.Context, w http.ResponseWriter, r *http.Request)
	GetAccessToken(ctx context.Context, r *http.Request) (string, error)
	GetCsrfFromCookie(ctx context.Context, r *http.Request) string
	GetCsrfFromHeader(r *http.Request) string
	ValidateSecFetchSite(r *http.Request) bool
}

type Csrf struct {
	secret          []byte
	randNonceLength int
}

func NewCsrf(secretKey string) (*Csrf, error) {
	log := logger.GetLogger()
	if secretKey == "" {
		log.Fatal("CSRF secret key is required")
		return &Csrf{}, CsrfSecretError
	}
	return &Csrf{
		secret:          []byte(secretKey),
		randNonceLength: 32,
	}, nil
}

func (obj *Csrf) GenerateRandValue(ctx context.Context) (string, error) {
	data := make([]byte, obj.randNonceLength)
	_, err := rand.Read(data)
	if err != nil {
		log := logger.GetLoggerWIthRequestId(ctx)
		log.Error("Error generating random value",
			zap.Error(err))
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func (obj *Csrf) GenerateCsrf(ctx context.Context, token string) (string, error) {
	randomValue, err := obj.GenerateRandValue(ctx)
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("%d!%s!%d!%s", len(token), token, len(randomValue), randomValue)
	hmacObj := hmac.New(sha256.New, obj.secret)
	hmacObj.Write([]byte(message))
	hmacValue := hex.EncodeToString(hmacObj.Sum(nil))
	csrfToken := hmacValue + "." + randomValue
	return csrfToken, nil
}

func (obj *Csrf) ValidateCsrf(ctx context.Context, csrf string, token string) (bool, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	parts := strings.Split(csrf, ".")
	if len(parts) != 2 {
		log.Warn("Get invalid CSRF token")
		return false, InvalidCsrfError
	}
	hmacValue := parts[0]
	randomValue := parts[1]
	message := fmt.Sprintf("%d!%s!%d!%s", len(token), token, len(randomValue), randomValue)
	hmacObj := hmac.New(sha256.New, obj.secret)
	hmacObj.Write([]byte(message))
	expectedHmac := hex.EncodeToString(hmacObj.Sum(nil))
	if !hmac.Equal([]byte(hmacValue), []byte(expectedHmac)) {
		log.Warn("Get wrong CSRF token")
		return false, InvalidCsrfSignatureError
	}
	return true, nil
}

func (obj *Csrf) GetCsrfValueFromToken(ctx context.Context, csrf string) (string, error) {
	parts := strings.Split(csrf, ".")
	if len(parts) != 2 {
		log := logger.GetLoggerWIthRequestId(ctx)
		log.Warn("Get invalid CSRF value")
		return "", InvalidCsrfError
	}
	return parts[1], nil
}

func (obj *Csrf) SetCsrfCookie(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	token, err := obj.GetAccessToken(ctx, r)
	if err != nil {
		return
	}
	signedCsrfToken, err := obj.GenerateCsrf(ctx, token)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CsrfCookieName, // ToDo: __Host-csrf_token
		Value:    signedCsrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(CsrfCookieExpirationTime),
	})
	rawCsrfToken, err := obj.GetCsrfValueFromToken(ctx, signedCsrfToken)
	if err != nil {
		return
	}
	w.Header().Set(CsrfHeaderName, rawCsrfToken)
}

func (obj *Csrf) GetAccessToken(ctx context.Context, r *http.Request) (string, error) {
	cookie, err := r.Cookie(auth.TokenName)
	if err != nil {
		log := logger.GetLoggerWIthRequestId(ctx)
		log.Warn("Failed to get access token",
			zap.Error(err))
		return "", err
	}
	return cookie.Value, nil
}

func (obj *Csrf) GetCsrfFromCookie(ctx context.Context, r *http.Request) string {
	cookie, err := r.Cookie(CsrfCookieName)
	if err != nil {
		log := logger.GetLoggerWIthRequestId(ctx)
		log.Warn("Failed to get CSRF cookie",
			zap.Error(err))
		return ""
	}
	return cookie.Value
}

func (obj *Csrf) GetCsrfFromHeader(r *http.Request) string {
	token := r.Header.Get(CsrfHeaderName)
	return token
}

func (obj *Csrf) ValidateSecFetchSite(r *http.Request) bool {
	log := logger.GetLoggerWIthRequestId(r.Context())
	secFetchSite := r.Header.Get("Sec-Fetch-Site")
	if secFetchSite == "" {
		log.Info("[CSRF middleware] Sec-Fetch-Site header missing, relying on CSRF tokens")
		return true
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	switch secFetchSite {
	case "same-origin":
		log.Info("[CSRF middleware] Sec-Fetch-Site: same-origin, request allowed")
		return true
	case "same-site":
		log.Info("[CSRF middleware] Sec-Fetch-Site: same-site, request allowed")
		return true
	case "cross-site":
		log.Warn("[CSRF middleware] Sec-Fetch-Site: cross-site, request blocked")
		return false
	case "none":
		log.Info("[CSRF middleware] Sec-Fetch-Site: none, request allowed")
		return true
	default:
		log.Warn("[CSRF middleware] Sec-Fetch-Site: unknown, request blocked",
			zap.String("Sec-Fetch-Site", secFetchSite))
		return false
	}
}
