package main

import (
	"fmt"
	"main/auth"
	"main/jwt"
	"net/http"
)

func root_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Привет, мир!</h1>")
	isAuth, userID := auth.IsAuth(r)
	fmt.Fprintln(w, "Is auth = ", isAuth)
	if isAuth {
		fmt.Fprintln(w, "user id = ", userID)
	}
}

func login_handler(w http.ResponseWriter, r *http.Request) {
	auth.GenerateNewAuthCookie(w, "1")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница логина</h1>")
}

func signup_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница регистрации</h1>")
}

func logout_handler(w http.ResponseWriter, r *http.Request) {
	auth.ClearOldToken(w, r)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница выхода</h1>")
}

func main() {
	jwt.TokenStore.NextID = 0
	http.HandleFunc("/", root_handler)
	http.HandleFunc("/login/", login_handler)
	http.HandleFunc("/signup/", signup_handler)
	http.HandleFunc("/logout/", logout_handler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
