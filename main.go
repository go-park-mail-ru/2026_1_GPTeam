package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	application2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/middleware"
	repository2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	web2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web"
	"github.com/jackc/pgx/v5"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer func() {
		err = conn.Close(context.Background())
		if err != nil {
			fmt.Printf("Unable to close connection: %v\n", err)
		}
	}()

	userRepo := repository2.NewPostgresUser(conn)
	budgetRepo := repository2.NewPostgresBudget(conn)
	jwtRepo := repository2.NewPostgresJwt(conn)

	userUseCases := application2.NewUser(userRepo)
	jwtUseCases, err := jwt_auth.NewJwt(jwtRepo, os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
	if err != nil {
		fmt.Println(err)
		return
	}
	authUseCases := auth.NewJWTAuth(jwtUseCases)
	budgetUseCases := application2.NewBudget(budgetRepo)

	userHandlers := web2.NewUserHandler(userUseCases)
	authHandlers := web2.NewJWTHandler(authUseCases, userUseCases)
	budgetHandlers := web2.NewBudgetHandler(budgetUseCases)

	mux := http.NewServeMux()
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandlers.Logout)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandlers.RefreshToken)))
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandlers.SignUp)))
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandlers.Login)))
	mux.Handle("/profile", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandlers.Profile)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandlers.Balance)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandlers.GetBudgets)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandlers.GetBudget)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(budgetHandlers.Create)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(budgetHandlers.Delete)))

	handler := middleware.AuthMiddleware(mux, authUseCases, userUseCases)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.PanicMiddleware(handler)

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
