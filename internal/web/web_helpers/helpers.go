package web_helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func WriteResponseJSON(w http.ResponseWriter, code int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func GetAuthUser(r *http.Request) (models.UserModel, bool) {
	user := r.Context().Value("user")
	if user == nil {
		return models.UserModel{}, false
	}

	if authUser, ok := user.(models.UserModel); ok {
		return authUser, true
	}

	if authUserPtr, ok := user.(*models.UserModel); ok && authUserPtr != nil {
		return *authUserPtr, true
	}

	fmt.Printf("user is a %T\n", user)
	return models.UserModel{}, false
}

func ReadRequestJSON(r *http.Request, request any) error {
	return json.NewDecoder(r.Body).Decode(request)
}

func SetCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONT_URL"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Origin, Cache-Control, X-Requested-With")
}
