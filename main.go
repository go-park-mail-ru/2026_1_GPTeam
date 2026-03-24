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
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := logger.InitLogger()
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		return
	}
	defer func() {
		err = logger.Close()
		if err != nil {
			fmt.Println("Error closing logger: ", err)
		}
	}()
	log := logger.GetLogger()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Failed to create pool: ", zap.Error(err))
		return
	}
	defer pool.Close()
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database: ", zap.Error(err))
	}

	enumsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	enumsPostgres, err := repository.NewEnumsPostgres(enumsCtx, pool)
	if err != nil {
		log.Fatal("Error get data from enums table: ", zap.Error(err))
		return
	}
	userPostgres := repository.NewUserPostgres(pool)
	budgetPostgres := repository.NewBudgetPostgres(pool)
	jwtPostgres := repository.NewJwtPostgres(pool)
	transactionPostgres := repository.NewTransactionPostgres(pool)
	accountPostgres := repository.NewAccountPostgres(pool)
	log.Info("repositories initialized")

	enumsApp := application.NewEnums(enumsPostgres)
	userApp := application.NewUser(userPostgres)
	jwt, err := jwt_auth.NewJwt(jwtPostgres, os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
	if err != nil {
		log.Fatal("Error creating JWT use cases: ", zap.Error(err))
		return
	}
	authService := auth.NewJwtAuthService(jwt)
	budgetApp := application.NewBudget(budgetPostgres)
	transactionApp := application.NewTransaction(transactionPostgres)
	accountApp := application.NewAccount(accountPostgres)
	log.Info("use cases initialized")

	enumsHandler := web.NewEnumsHandler(enumsApp)
	userHandler := web.NewUserHandler(userApp)
	authHandler := web.NewAuthHandler(authService, userApp, accountApp)
	budgetHandler := web.NewBudgetHandler(budgetApp, enumsApp)
	transactionHandler := web.NewTransactionHandler(transactionApp, enumsApp, accountApp)
	accountHandler := web.NewAccountHandler(accountApp)
	log.Info("handlers initialized")

	mux := http.NewServeMux()
	mux.Handle("/account", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(accountHandler.GetAccount)))
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.SignUp)))
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/profile", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPatch)(http.HandlerFunc(userHandler.ProfileHandler)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.Balance)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudgets)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudget)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(budgetHandler.Delete)))
	mux.Handle("/transactions", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(transactionHandler.Transactions)))
	mux.Handle("/transactions/{id}", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodDelete)(http.HandlerFunc(transactionHandler.Transaction)))
	mux.Handle("/enums/get_currency_codes", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CurrencyCodes)))
	mux.Handle("/enums/get_transaction_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.TransactionTypes)))
	mux.Handle("/enums/get_category_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CategoryTypes)))
	handler := middleware.AuthMiddleware(mux, authService, userApp)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.PanicMiddleware(handler)

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Info("starting server at :8080")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server: ", zap.Error(err))
		return
	}
}
