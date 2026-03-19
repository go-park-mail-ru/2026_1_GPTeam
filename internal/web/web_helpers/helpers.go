package web_helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func WriteResponseJSON(w http.ResponseWriter, code int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func SetCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONT_URL"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Origin, Cache-Control, X-Requested-With")
}
