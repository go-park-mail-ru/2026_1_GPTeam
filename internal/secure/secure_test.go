package secure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestSanitizeXss(t *testing.T) {
	t.Parallel()
	XssSanitizerInit()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Пустая строка",
			input:    "",
			expected: "",
		},
		{
			name:     "Безопасная строка",
			input:    "Обычный текст 123",
			expected: "Обычный текст 123",
		},
		{
			name:     "Строка с XSS",
			input:    `<script>alert("hack")</script>Обычный текст`,
			expected: "Обычный текст",
		},
		{
			name:     "Строка с HTML тегами",
			input:    "<b>Жирный</b> и <i>курсив</i>",
			expected: "Жирный и курсив",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeXss(c.input)
			require.Equal(t, c.expected, result)
		})
	}
}

func getTestContext() context.Context {
	return context.WithValue(context.Background(), "request_id", "test-req-id")
}

func TestNewCsrf(t *testing.T) {
	t.Parallel()

	t.Run("Успешное создание", func(t *testing.T) {
		t.Parallel()
		csrfService, err := NewCsrf("my-super-secret-key")
		require.NoError(t, err)
		require.NotNil(t, csrfService)
		require.Equal(t, "my-super-secret-key", string(csrfService.secret))
		require.Equal(t, 32, csrfService.randNonceLength)
	})
}

func TestCsrf_GenerateAndValidate(t *testing.T) {
	t.Parallel()
	ctx := getTestContext()
	csrfService, _ := NewCsrf("my-super-secret-key")
	token := "access-token-123"

	t.Run("Успешная генерация и валидация", func(t *testing.T) {
		csrfValue, err := csrfService.generateCsrf(ctx, token)
		require.NoError(t, err)
		require.NotEmpty(t, csrfValue)

		isValid, err := csrfService.ValidateCsrf(ctx, csrfValue, token)
		require.NoError(t, err)
		require.True(t, isValid)
	})

	t.Run("Валидация с неверным форматом (без точки)", func(t *testing.T) {
		isValid, err := csrfService.ValidateCsrf(ctx, "invalid_format_token", token)
		require.ErrorIs(t, err, InvalidCsrfError)
		require.False(t, isValid)
	})

	t.Run("Валидация с неверной сигнатурой", func(t *testing.T) {
		csrfValue, _ := csrfService.generateCsrf(ctx, token)
		tamperedCsrf := "tamperedHMAC" + csrfValue[6:]

		isValid, err := csrfService.ValidateCsrf(ctx, tamperedCsrf, token)
		require.ErrorIs(t, err, InvalidCsrfSignatureError)
		require.False(t, isValid)
	})

	t.Run("Валидация с другим Access Token", func(t *testing.T) {
		csrfValue, _ := csrfService.generateCsrf(ctx, token)

		isValid, err := csrfService.ValidateCsrf(ctx, csrfValue, "another-token")
		require.ErrorIs(t, err, InvalidCsrfSignatureError)
		require.False(t, isValid)
	})
}

func TestCsrf_GetCsrfValueFromToken(t *testing.T) {
	t.Parallel()
	ctx := getTestContext()
	csrfService, _ := NewCsrf("secret")

	t.Run("Успешное получение", func(t *testing.T) {
		val, err := csrfService.getCsrfValueFromToken(ctx, "hmac.randomPart")
		require.NoError(t, err)
		require.Equal(t, "randomPart", val)
	})

	t.Run("Неверный формат", func(t *testing.T) {
		val, err := csrfService.getCsrfValueFromToken(ctx, "invalid")
		require.ErrorIs(t, err, InvalidCsrfError)
		require.Empty(t, val)
	})
}

func TestCsrf_SetCsrfCookie(t *testing.T) {
	t.Parallel()
	ctx := getTestContext()
	csrfService, _ := NewCsrf("secret")

	t.Run("Нет Access Token — прерывается без установки", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)

		csrfService.SetCsrfCookie(ctx, w, req)

		require.Empty(t, w.Result().Cookies())
		require.Empty(t, w.Header().Get(CsrfHeaderName))
	})

	t.Run("Успешная установка куки и заголовка", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		req.AddCookie(&http.Cookie{Name: auth.TokenName, Value: "valid-token"})

		csrfService.SetCsrfCookie(ctx, w, req)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		require.Equal(t, CsrfCookieName, cookies[0].Name)
		require.NotEmpty(t, cookies[0].Value)

		headerVal := w.Header().Get(CsrfHeaderName)
		require.NotEmpty(t, headerVal)

		valFromCookie, _ := csrfService.getCsrfValueFromToken(ctx, cookies[0].Value)
		require.Equal(t, valFromCookie, headerVal)
	})
}

func TestCsrf_GetAccessToken(t *testing.T) {
	t.Parallel()
	ctx := getTestContext()
	csrfService, _ := NewCsrf("secret")

	t.Run("Токен есть", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: auth.TokenName, Value: "test-token"})

		val, err := csrfService.GetAccessToken(ctx, req)
		require.NoError(t, err)
		require.Equal(t, "test-token", val)
	})

	t.Run("Токена нет", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		val, err := csrfService.GetAccessToken(ctx, req)
		require.ErrorIs(t, err, http.ErrNoCookie)
		require.Empty(t, val)
	})
}

func TestCsrf_GetCsrfFromCookie(t *testing.T) {
	t.Parallel()
	ctx := getTestContext()
	csrfService, _ := NewCsrf("secret")

	t.Run("Кука есть", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: CsrfCookieName, Value: "cookie-csrf-token"})

		val := csrfService.GetCsrfFromCookie(ctx, req)
		require.Equal(t, "cookie-csrf-token", val)
	})

	t.Run("Куки нет", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		val := csrfService.GetCsrfFromCookie(ctx, req)
		require.Empty(t, val)
	})
}

func TestCsrf_GetCsrfFromHeader(t *testing.T) {
	t.Parallel()
	csrfService, _ := NewCsrf("secret")

	t.Run("Хэдер есть", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(CsrfHeaderName, "header-csrf-token")

		val := csrfService.GetCsrfFromHeader(req)
		require.Equal(t, "header-csrf-token", val)
	})

	t.Run("Хэдера нет", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		val := csrfService.GetCsrfFromHeader(req)
		require.Empty(t, val)
	})
}

func TestCsrf_ValidateSecFetchSite(t *testing.T) {
	t.Parallel()
	csrfService, _ := NewCsrf("secret")

	cases := []struct {
		name       string
		method     string
		secFetch   string
		wantResult bool
	}{
		{"Пустой заголовок -> true", http.MethodPost, "", true},
		{"Метод GET -> true", http.MethodGet, "cross-site", true},
		{"Метод HEAD -> true", http.MethodHead, "cross-site", true},
		{"Метод OPTIONS -> true", http.MethodOptions, "cross-site", true},
		{"same-origin -> true", http.MethodPost, "same-origin", true},
		{"same-site -> true", http.MethodPost, "same-site", true},
		{"none -> true", http.MethodPost, "none", true},
		{"cross-site -> false", http.MethodPost, "cross-site", false},
		{"Неизвестное значение -> false", http.MethodPost, "unknown-value", false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/", nil)
			if c.secFetch != "" {
				req.Header.Set("Sec-Fetch-Site", c.secFetch)
			}
			req = req.WithContext(getTestContext())

			result := csrfService.ValidateSecFetchSite(req)
			require.Equal(t, c.wantResult, result)
		})
	}
}
