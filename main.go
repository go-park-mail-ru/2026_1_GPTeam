package main

import (
	"fmt"
	"main/auth"
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
	cookie, err := auth.GenerateNewAuthCookie("1")
	if err != nil {
		return
	}
	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница логина</h1>")
	fmt.Fprintln(w, cookie.Value)
}

func signup_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница регистрации</h1>")
}

func logout_handler(w http.ResponseWriter, r *http.Request) {
	auth.ClearOldToken(w)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница выхода</h1>")
}

func main() {
	http.HandleFunc("/", root_handler)
	http.HandleFunc("/login/", login_handler)
	http.HandleFunc("/signup/", signup_handler)
	http.HandleFunc("/logout/", logout_handler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
