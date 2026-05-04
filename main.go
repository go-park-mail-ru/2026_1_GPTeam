package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	fileupload "github.com/go-park-mail-ru/2026_1_GPTeam/internal/clients/fileserver"
	groq "github.com/go-park-mail-ru/2026_1_GPTeam/internal/clients/groq"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure/rate_limiter"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web"
	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtVersion := os.Getenv("JWT_VERSION")
	if len(strings.TrimSpace(jwtSecret)) < 8 {
		log.Fatal("JWT_SECRET must be at least 8 characters")
	}
	if jwtVersion == "" {
		log.Fatal("JWT_VERSION is required")
	}

	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}
	authConn, err := grpc.NewClient(
		authGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "tcp4", addr)
		}),
	)
	if err != nil {
		log.Fatal("auth gRPC dial", zap.String("addr", authGrpcAddr), zap.Error(err))
		return
	}
	defer authConn.Close()
	authClient := authv1.NewAuthServiceClient(authConn)
	authService := auth.NewGrpcAuthAdapter(authClient, jwtSecret, jwtVersion)

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
	transactionPostgres := repository.NewTransactionPostgres(pool)
	accountPostgres := repository.NewAccountPostgres(pool)
	supportPostgres := repository.NewPostgresSupport(pool)
	log.Info("repositories initialized")

	fsGrpcAddr := strings.TrimSpace(os.Getenv("FILESERVER_GRPC_ADDR"))
	if fsGrpcAddr == "" {
		fsGrpcAddr = "localhost:50053"
	}
	fsConn, err := grpc.NewClient(
		fsGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(8<<20),
			grpc.MaxCallRecvMsgSize(8<<20),
		),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "tcp4", addr)
		}),
	)
	if err != nil {
		log.Fatal("fileserver gRPC dial", zap.String("addr", fsGrpcAddr), zap.Error(err))
		return
	}
	defer fsConn.Close()
	fsClient := fsv1.NewFileServiceClient(fsConn)
	var avatarUploader application.AvatarUploader = fileupload.NewGrpcUploader(fsClient)

	enumsApp := application.NewEnums(enumsPostgres)
	userApp := application.NewUser(userPostgres, enumsApp, avatarUploader)
	csrfService, err := secure.NewCsrf(os.Getenv("CSRF_SECRET"))
	if err != nil {
		return
	}
	transactionApp := application.NewTransaction(transactionPostgres, accountPostgres)
	accountApp := application.NewAccount(accountPostgres)
	budgetApp := application.NewBudget(budgetPostgres, transactionApp, accountApp)
	supportApp := application.NewSupport(supportPostgres)
	groqClient := groq.NewGroqClient(groqKey, proxyURLStr)
	voiceApp := application.NewVoiceTransactionService(groqClient, enumsApp)
	log.Info("use cases initialized")

	enumsHandler := web.NewEnumsHandler(enumsApp)
	userHandler := web.NewUserHandler(userApp, accountApp)
	authHandler := web.NewAuthHandler(authService, userApp, accountApp)
	budgetHandler := web.NewBudgetHandler(budgetApp, enumsApp)
	transactionHandler := web.NewTransactionHandler(transactionApp, enumsApp, accountApp)
	accountHandler := web.NewAccountHandler(accountApp)
	voiceHandler := web.NewVoiceHandler(voiceApp, enumsApp)
	supportHandler := web.NewSupportHandler(supportApp, userApp)
	log.Info("handlers initialized")

	secure.XssSanitizerInit()
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisUrl := fmt.Sprintf("redis://%s:%s/0", redisHost, redisPort)
	redisPool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.DialURL(redisUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis: %w", err)
			}
			return conn, nil
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < 30*time.Second {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
	defer func() {
		err = redisPool.Close()
		if err != nil {
			fmt.Println("redis pool close error", err)
		}
	}()
	rateLimitBucket := repository.NewBucketRedis(redisPool)
	rateLimiter, err := rate_limiter.NewRateLimiter(rateLimitBucket, os.Getenv("SERVER_IP"))
	if err != nil {
		return
	}
	log.Info("secure package initialized")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.SignUp)))
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/api/account", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(accountHandler.GetAccount)))
	mux.Handle("/api/accounts", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(accountHandler.GetAccounts)))
	mux.Handle("/api/profile", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPatch)(http.HandlerFunc(userHandler.ProfileHandler)))
	mux.Handle("/api/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.Balance)))
	mux.Handle("/api/profile/avatar", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(userHandler.UploadAvatar)))
	mux.Handle("/transactions", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(transactionHandler.Transactions)))
	mux.Handle("/api/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudgets)))
	mux.Handle("/api/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(budgetHandler.GetBudget)))
	mux.Handle("/api/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("/api/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(budgetHandler.Delete)))
	mux.Handle("/api/transactions", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(transactionHandler.Transactions)))
	mux.Handle("/api/transactions/{id}", middleware.MethodValidationMiddleware(http.MethodGet, http.MethodDelete, http.MethodPut)(http.HandlerFunc(transactionHandler.Transaction)))
	mux.Handle("/api/transactions/search", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(transactionHandler.Search)))
	mux.Handle("/api/transactions/voice", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(voiceHandler.CreateVoiceTransaction)))
	mux.Handle("/enums/get_currency_codes", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CurrencyCodes)))
	mux.Handle("/enums/get_transaction_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.TransactionTypes)))
	mux.Handle("/enums/get_category_types", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(enumsHandler.CategoryTypes)))
	mux.Handle("/support/get_all_appeals", middleware.MethodValidationMiddleware(http.MethodGet)(middleware.OnlyStaffMiddleware(http.HandlerFunc(supportHandler.GetAll), userApp)))
	mux.Handle("/support/get_appeal/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(supportHandler.Detail)))
	mux.Handle("/support/get_appeals", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(supportHandler.GetAllByUser)))
	mux.Handle("/support/create_appeal", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(supportHandler.Create)))
	mux.Handle("/support/update/{id}", middleware.MethodValidationMiddleware(http.MethodPut)(middleware.OnlyStaffMiddleware(http.HandlerFunc(supportHandler.Update), userApp)))
	mux.Handle("/api/is_staff", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(userHandler.IsStaff)))

	handler := middleware.CSPMiddleware(mux)
	handler = middleware.CSRFMiddleware(handler, csrfService)
	handler = middleware.AuthMiddleware(handler, authService, userApp)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RateLimitMiddleware(handler, rateLimiter)
	handler = middleware.AccessLogMiddleware(handler)
	handler = middleware.PanicMiddleware(handler)

	addr := ":" + os.Getenv("PORT")
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	log.Info("starting server",
		zap.String("addr", addr),
		zap.String("auth_grpc", authGrpcAddr),
		zap.String("fileserver_grpc", fsGrpcAddr),
	)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server", zap.Error(err))
		return
	}
}
