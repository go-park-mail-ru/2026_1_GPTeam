package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	cases := []struct {
		method   string
		wantCode int
	}{
		{http.MethodGet, http.StatusNoContent},
		{http.MethodPost, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
	}
	for _, c := range cases {
		c := c
		t.Run(c.method, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/healthz", nil)
			rr := httptest.NewRecorder()
			Healthz(rr, req)
			require.Equal(t, c.wantCode, rr.Code)
		})
	}
}

func TestNoDirListing_BlocksRoot(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/img/", nil)
	rr := httptest.NewRecorder()
	NoDirListing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestNoDirListing_PassesNonRoot(t *testing.T) {
	t.Parallel()

	called := false
	req := httptest.NewRequest(http.MethodGet, "/img/foo.png", nil)
	rr := httptest.NewRecorder()
	NoDirListing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestRouter_ServesImage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "pic.png")
	require.NoError(t, os.WriteFile(file, []byte("PNGDATA"), 0o644))

	srv := httptest.NewServer(NewRouter(dir))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/img/pic.png")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp2, err := http.Get(srv.URL + "/img/")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp2.Body.Close() })
	require.Equal(t, http.StatusNotFound, resp2.StatusCode)

	resp3, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp3.Body.Close() })
	require.Equal(t, http.StatusNoContent, resp3.StatusCode)
}
