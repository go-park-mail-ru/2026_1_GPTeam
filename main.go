package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web"
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

	enumsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	enumsPostgres, err := repository.NewEnumsPostgres(enumsCtx, conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	userPostgres := repository.NewUserPostgres(conn)
	budgetPostgres := repository.NewBudgetPostgres(conn)
	jwtPostgres := repository.NewJwtPostgres(conn)
	transactionPostgres := repository.NewTransactionPostgres(conn)
	accountPostgres := repository.NewAccountPostgres(conn)

	enumsApp := application.NewEnums(enumsPostgres)
	userApp := application.NewUser(userPostgres)
	jwt, err := jwt_auth.NewJwt(jwtPostgres, os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
	if err != nil {
		fmt.Println(err)
		return
	}
	authService := auth.NewJwtAuthService(jwt)
	budgetApp := application.NewBudget(budgetPostgres)
	transactionApp := application.NewTransaction(transactionPostgres)
	accountApp := application.NewAccount(accountPostgres)

	enumsHandler := web.NewEnumsHandler(enumsApp)
	userHandler := web.NewUserHandler(userApp)
	authHandler := web.NewAuthHandler(authService, userApp, accountApp)
	budgetHandler := web.NewBudgetHandler(budgetApp, enumsApp)
	transactionHandler := web.NewTransactionHandler(transactionApp, enumsApp, accountApp)

	mux := http.NewServeMux()
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.SignUp)))
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/profile", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.Profile)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.Balance)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudgets)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudget)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(budgetHandler.Delete)))
	mux.Handle("/enums/get_currency_codes", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CurrencyCodes)))
	mux.Handle("/transactions", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(transactionHandler.Transactions)))
	mux.Handle("/transactions/{id}", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodDelete, http.MethodPost)(http.HandlerFunc(transactionHandler.Transaction)))

	handler := middleware.AuthMiddleware(mux, authService, userApp)
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
