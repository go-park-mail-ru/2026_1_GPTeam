package httpserver

import (
	"net/http"
	"strings"
)

func NoDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path == "/img/" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewRouter(storageRoot string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", Healthz)
	fs := http.StripPrefix("/img/", http.FileServer(http.Dir(storageRoot)))
	mux.Handle("/img/", NoDirListing(fs))
	return mux
}
