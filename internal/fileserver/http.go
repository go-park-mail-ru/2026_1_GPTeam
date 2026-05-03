package fileserver

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const MaxInternalUploadBody = 6 << 20

func StripBearer(h string) (string, bool) {
	h = strings.TrimSpace(h)
	const p = "Bearer "
	if len(h) <= len(p) || !strings.EqualFold(h[:len(p)], p) {
		return "", false
	}
	return strings.TrimSpace(h[len(p):]), true
}

func NoDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path == "/img/" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func InternalUpload(storageRoot string, uploadToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		got, ok := StripBearer(r.Header.Get("Authorization"))
		if !ok || subtle.ConstantTimeCompare([]byte(got), []byte(uploadToken)) != 1 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if err := r.ParseMultipartForm(MaxInternalUploadBody); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		defer file.Close()

		ext := strings.TrimSpace(r.FormValue("extension"))
		if ext != "" && !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if ext == "" {
			ext = ".bin"
		}

		if err := os.MkdirAll(storageRoot, 0755); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		baseName := filepath.Base(uuid.New().String() + ext)
		path := filepath.Join(storageRoot, baseName)

		dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, io.LimitReader(file, MaxInternalUploadBody)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"filename": baseName})
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func InternalUploadDisabled(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
}

func NewRouter(storageRoot, uploadToken string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	if uploadToken != "" {
		mux.HandleFunc("/internal/upload", InternalUpload(storageRoot, uploadToken))
	} else {
		mux.HandleFunc("/internal/upload", InternalUploadDisabled)
	}
	fs := http.StripPrefix("/img/", http.FileServer(http.Dir(storageRoot)))
	mux.Handle("/img/", NoDirListing(fs))
	return mux
}
