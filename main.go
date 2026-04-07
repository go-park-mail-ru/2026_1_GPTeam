package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	groq "github.com/go-park-mail-ru/2026_1_GPTeam/internal/clients/groq"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file: ", err)
		return
	}
	DEBUG := os.Getenv("DEBUG") == "true"

	err = logger.InitLogger(DEBUG)
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

	err = logger.InitAccessLogger()
	if err != nil {
		log.Fatal("Error initializing access logger",
			zap.Error(err))
	}
	defer func() {
		err = logger.AccessClose()
		if err != nil {
			fmt.Println("Error closing access logger: ", err)
		}
	}()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", zap.Error(err))
		return
	}

	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	if groqKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
		return
	}
	proxyURLStr := os.Getenv("PROXY_URL")

	log.Info("groq api key loaded",
		zap.Int("len", len(groqKey)),
		zap.String("prefix", groqKey[:min(8, len(groqKey))]+"..."),
		zap.String("suffix", "..."+groqKey[max(0, len(groqKey)-4):]),
		zap.String("proxy", proxyURLStr),
	)

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Failed to create pool", zap.Error(err))
		return
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	enumsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	enumsPostgres, err := repository.NewEnumsPostgres(enumsCtx, pool)
	if err != nil {
		log.Fatal("Failed to create enums repo", zap.Error(err))
		return
	}
	userPostgres := repository.NewUserPostgres(pool)
	budgetPostgres := repository.NewBudgetPostgres(pool)
	jwtPostgres := repository.NewJwtPostgres(pool)
	transactionPostgres := repository.NewTransactionPostgres(pool)
	accountPostgres := repository.NewAccountPostgres(pool)
	log.Info("repositories initialized")

	enumsApp := application.NewEnums(enumsPostgres)
	userApp := application.NewUser(userPostgres, enumsApp)
	jwtService, err := jwt_auth.NewJwt(jwtPostgres, os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
	if err != nil {
		log.Fatal("Failed to create JWT service", zap.Error(err))
		return
	}
	authService := auth.NewJwtAuthService(jwtService)
	csrfService, err := secure.NewCsrf(os.Getenv("CSRF_SECRET"))
	if err != nil {
		return
	}
	budgetApp := application.NewBudget(budgetPostgres)
	transactionApp := application.NewTransaction(transactionPostgres)
	accountApp := application.NewAccount(accountPostgres)

	groqClient := groq.NewGroqClient(groqKey, proxyURLStr)
	voiceApp := application.NewVoiceTransactionService(groqClient, enumsApp)

	log.Info("use cases initialized")

	enumsHandler := web.NewEnumsHandler(enumsApp)
	userHandler := web.NewUserHandler(userApp)
	authHandler := web.NewAuthHandler(authService, userApp, accountApp)
	budgetHandler := web.NewBudgetHandler(budgetApp, enumsApp)
	transactionHandler := web.NewTransactionHandler(transactionApp, enumsApp, accountApp)
	accountHandler := web.NewAccountHandler(accountApp)
	voiceHandler := web.NewVoiceHandler(voiceApp, enumsApp)
	log.Info("handlers initialized")

	fileServer := http.StripPrefix("/img/", http.FileServer(http.Dir("./static")))

	secure.XssSanitizerInit()
	log.Info("secure package initialized")

	mux := http.NewServeMux()
	mux.Handle("/account", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(accountHandler.GetAccount)))
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.SignUp)))
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/profile", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPatch)(http.HandlerFunc(userHandler.ProfileHandler)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.Balance)))
	mux.Handle("/api/profile/avatar", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(userHandler.UploadAvatar)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudgets)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudget)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(budgetHandler.Delete)))
	mux.Handle("/transactions", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(transactionHandler.Transactions)))
	mux.Handle("/transactions/voice", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(voiceHandler.CreateVoiceTransaction)))
	mux.Handle("/transactions/{id}", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodDelete, http.MethodPut)(http.HandlerFunc(transactionHandler.Transaction)))
	mux.Handle("/enums/get_currency_codes", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CurrencyCodes)))
	mux.Handle("/enums/get_transaction_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.TransactionTypes)))
	mux.Handle("/enums/get_category_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CategoryTypes)))
	mux.Handle("/img/", middleware.NoDirListing(fileServer))

	handler := middleware.CSRFMiddleware(mux, csrfService)
	handler = middleware.AuthMiddleware(handler, authService, userApp)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.AccessLogMiddleware(handler)
	handler = middleware.PanicMiddleware(handler)

	addr := ":" + os.Getenv("PORT")
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	log.Info("starting server", zap.String("addr", addr))
	if DEBUG {
		err = server.ListenAndServe()
	} else {
		cerfFile := os.Getenv("CERT_FILE")
		if cerfFile == "" {
			log.Fatal("CERT_FILE not set")
			return
		}
		keyFile := os.Getenv("KEY_FILE")
		if keyFile == "" {
			log.Fatal("KEY_FILE not set")
			return
		}
		err = server.ListenAndServeTLS(cerfFile, keyFile)
	}
	if err != nil {
		log.Fatal("Error starting server", zap.Error(err))
		return
	}
}
