package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	aiprovider "booknest/internal/ai/provider"
	"booknest/internal/domain"
	"booknest/internal/http/api"
	apiv1 "booknest/internal/http/api/v1"
	"booknest/internal/http/controller"
	"booknest/internal/http/database"
	"booknest/internal/middleware"
	"booknest/internal/pkg/util"
	"booknest/internal/repository"
	"booknest/internal/service/ai_service"
	"booknest/internal/service/author_service"
	"booknest/internal/service/book_embedding_service"
	"booknest/internal/service/book_service"
	"booknest/internal/service/cart_service"
	"booknest/internal/service/category_service"
	"booknest/internal/service/notification_service"
	"booknest/internal/service/order_service"
	"booknest/internal/service/publisher_service"
	"booknest/internal/service/review_service"
	"booknest/internal/service/user_service"
)

var connectGORM = database.ConnectGORM

type orderServiceRuntimeConfig struct {
	UseMicroservice bool
	GRPCAddress     string
}

func loadOrderServiceRuntimeConfig() (orderServiceRuntimeConfig, error) {
	// Get gRPC address from env file and trim spaces
	cfg := orderServiceRuntimeConfig{
		// Get gRPC address from env file
		GRPCAddress: strings.TrimSpace(os.Getenv("ORDER_SERVICE_GRPC_ADDR")),
	}

	// Check ORDER_SERVICE_MODE first for backward compatibility
	if mode := strings.ToLower(strings.TrimSpace(os.Getenv("ORDER_SERVICE_MODE"))); mode != "" {
		switch mode {
		case "monolith":
			cfg.UseMicroservice = false
		case "microservice":
			cfg.UseMicroservice = true
		default:
			return orderServiceRuntimeConfig{}, fmt.Errorf("invalid ORDER_SERVICE_MODE %q: use monolith or microservice", mode)
		}
	}

	// Then check USE_ORDER_MICROSERVICE for backward compatibility with older env files
	if raw := strings.TrimSpace(os.Getenv("USE_ORDER_MICROSERVICE")); raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return orderServiceRuntimeConfig{}, fmt.Errorf("parse USE_ORDER_MICROSERVICE: %w", err)
		}
		cfg.UseMicroservice = enabled
	}

	// If microservice mode is enabled, GRPCAddress must be set
	if cfg.UseMicroservice && cfg.GRPCAddress == "" {
		return orderServiceRuntimeConfig{}, errors.New("ORDER_SERVICE_GRPC_ADDR is required when the order microservice is enabled")
	}

	// If microservice mode is disabled, GRPCAddress is ignored but we can log it if set
	return cfg, nil
}

func initRedisClient() (*redis.Client, error) {
	// Get Redis host address
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		return nil, nil
	}

	// Get DB from env file
	db := 0
	if rawDB := strings.TrimSpace(os.Getenv("REDIS_DB")); rawDB != "" {
		parsed, err := strconv.Atoi(rawDB)
		if err != nil {
			return nil, fmt.Errorf("parse REDIS_DB: %w", err)
		}
		db = parsed
	}

	var tlsConfig *tls.Config

	if os.Getenv("ENV") == "production" {
		tlsConfig = &tls.Config{}
	}
	// Get a new client
	client := redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  os.Getenv("REDIS_PASSWORD"),
		DB:        db,
		TLSConfig: tlsConfig,
	})

	// Ping the redis DB
	if err := client.Ping().Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis at %s: %w", addr, err)
	}

	return client, nil
}

func useCORSMiddleware(allowedOrigins map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization",
			)
			c.Writer.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, OPTIONS",
			)
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func frontendAllowedOrigins() map[string]bool {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:5173": true,
	}

	for _, envName := range []string{"BOOKNEST_WEB_URL", "FRONTEND_URL"} {
		origin := strings.TrimRight(strings.TrimSpace(os.Getenv(envName)), "/")
		if origin != "" {
			allowedOrigins[origin] = true
		}
	}

	return allowedOrigins
}

func initNotificationService(
	ctx context.Context,
	notificationRepo domain.NotificationRepository,
) (domain.NotificationService, error) {
	from := strings.TrimSpace(os.Getenv("SES_FROM_EMAIL"))
	if from == "" {
		slog.Info("SES_FROM_EMAIL not set, using EMAIL_FROM")
		from = strings.TrimSpace(os.Getenv("EMAIL_FROM"))
	}

	sesRegion := strings.TrimSpace(os.Getenv("SES_REGION"))
	sesAccessKey := strings.TrimSpace(os.Getenv("SES_ACCESS_KEY"))
	sesSecretKey := strings.TrimSpace(os.Getenv("SES_SECRET_KEY"))

	awsAccessKey := strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	awsSecretKey := strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	awsRegion := strings.TrimSpace(os.Getenv("AWS_REGION"))

	var emailProvider domain.EmailProvider
	if from != "" && sesRegion != "" && sesAccessKey != "" && sesSecretKey != "" {
		cfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(sesRegion),
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(sesAccessKey, sesSecretKey, ""),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("load email notification config: %w", err)
		}
		emailProvider = notification_service.NewSESEmail(ses.NewFromConfig(cfg), from)
	} else {
		slog.Warn("SES email notifications are not fully configured")
	}

	var smsProvider domain.SMSProvider
	if awsRegion != "" && awsAccessKey != "" && awsSecretKey != "" {
		cfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(awsRegion),
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, ""),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("load sms notification config: %w", err)
		}
		smsProvider = notification_service.NewSNSSMS(sns.NewFromConfig(cfg))
	} else {
		slog.Warn("SNS SMS notifications are not fully configured")
	}

	if emailProvider == nil && smsProvider == nil {
		return nil, fmt.Errorf("notifications are not configured: provide SES and/or SNS credentials")
	}

	return notification_service.NewNotificationServiceWithProvidersAndRepository(
		emailProvider,
		smsProvider,
		notificationRepo,
	), nil
}

func SetupServer(dbpool *pgxpool.Pool) (*gin.Engine, error) {
	configureSwagger()

	gormdb, err := connectGORM()
	if err != nil {
		return nil, fmt.Errorf("connect gorm: %w", err)
	}

	sqlDB, err := gormdb.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db handle: %w", err)
	}

	jwtConfig, err := middleware.LoadJWTConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("load jwt config: %w", err)
	}
	controller.SetJWTConfig(jwtConfig)

	// Initialise redis
	redisClient, err := initRedisClient()
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	// Add redis client
	controller.SetRedisClient(redisClient)
	if redisClient == nil {
		log.Println("redis not configured (REDIS_ADDR empty); using in-memory login rate limiter")
	}

	userRepo := repository.NewUserRepo(dbpool, gormdb)
	vtRepo := repository.NewVerificationRepo(dbpool, gormdb)
	notificationRepo := repository.NewNotificationRepo(dbpool)
	txm := util.NewTransactionManager(dbpool)

	// Initialise notification service with SES email provider
	notificationService, err := initNotificationService(context.Background(), notificationRepo)
	if err != nil {
		return nil, err
	}

	userService := user_service.NewUserServiceWithNotification(txm, userRepo, vtRepo, notificationService)
	userController := controller.NewUserController(userService)

	bookRepo := repository.NewBookRepository(gormdb, sqlDB)
	bookEmbeddingRepo := repository.NewBookEmbeddingRepository(gormdb, sqlDB)

	authorRepo := repository.NewAuthorRepo(gormdb)
	authorService := author_service.NewAuthorService(authorRepo)
	authorController := controller.NewAuthorController(authorService)

	categoryRepo := repository.NewCategoryRepo(gormdb)
	categoryService := category_service.NewCategoryService(categoryRepo)
	categoryController := controller.NewCategoryController(categoryService)

	publisherRepo := repository.NewPublisherRepo(dbpool, gormdb)
	publisherService := publisher_service.NewPublisherService(publisherRepo)
	publisherController := controller.NewPublisherController(publisherService)

	cartRepo := repository.NewCartRepo(dbpool)
	cartService := cart_service.NewCartService(cartRepo, bookRepo)
	cartController := controller.NewCartController(cartService)

	orderRepo := repository.NewOrderRepo(dbpool)
	orderService := order_service.NewOrderServiceWithNotification(txm, orderRepo, cartRepo, userRepo, notificationService)

	// Load order config from env variables to determine whether to use microservice or monolith implementation
	orderConfig, err := loadOrderServiceRuntimeConfig()
	if err != nil {
		return nil, fmt.Errorf("load order service config: %w", err)
	}

	// If microservice mode is enabled,
	// wrap the monolith order service with a gRPC client that forwards calls to the microservice
	if orderConfig.UseMicroservice {
		orderService, err = order_service.NewRemoteOrderService(
			orderConfig.GRPCAddress,
			orderService,
			cartRepo,
			bookRepo,
			userRepo,
			notificationService,
		)
		if err != nil {
			return nil, fmt.Errorf("init remote order service: %w", err)
		}
		slog.Info("using order microservice", "addr", orderConfig.GRPCAddress)
	} else {
		slog.Info("using monolith order service")
	}
	orderController := controller.NewOrderController(orderService)

	reviewRepo := repository.NewReviewRepository(gormdb)
	reviewService := review_service.NewReviewService(reviewRepo, orderRepo)
	reviewController := controller.NewReviewController(reviewService)
	imageController := controller.NewImageController()

	// Initialise AI provider, service, and controller
	aiProvider, err := aiprovider.NewProviderFromEnv()
	if err != nil {
		slog.Warn("AI provider is not configured", "error", err)
	}
	aiService := ai_service.NewAIService(aiProvider)
	aiController := controller.NewAIController(aiService)

	// Initialise Book service after AI (best-effort summary generation on create/update).
	embeddingService := book_embedding_service.New(aiService, bookEmbeddingRepo)
	bookService := book_service.NewBookService(bookRepo, categoryRepo, embeddingService, bookEmbeddingRepo, orderRepo, aiService)
	bookController := controller.NewBookController(bookService)

	r := gin.Default()
	r.Use(useCORSMiddleware(frontendAllowedOrigins()))
	r.Use(middleware.SecurityHeaders())
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.ErrorHandler())
	r.GET(
		"/swagger/v1/*any",
		middleware.SwaggerAuthMiddleware(),
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.InstanceName(swaggerV1InstanceName),
			ginSwagger.URL(swaggerV1DocURL),
		),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1Router := apiv1.NewRouter(
		userController,
		bookController,
		authorController,
		categoryController,
		publisherController,
		cartController,
		orderController,
		reviewController,
		imageController,
		aiController,
	)

	// Mount only v1 now; v2 can be plugged in with another registrar later.
	api.MountVersions(r, v1Router)

	return r, nil
}

// StartHTTPServer starts the HTTP server — only used by main.go
func StartHTTPServer(r *gin.Engine) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// If we don’t run it in a goroutine, shutdown logic will never execute
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Println("🚀 BookNest backend started on http://localhost:8080")

	// graceful shutdown
	/*
		* Creates a channel to receive OS signals and Listens for:
			- Ctrl + C
			- Docker stop
			- Pod termination
			<-quit blocks until signal arrives
	*/
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Gives active requests 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	/*
		Why server shut down:
		1. Stops accepting new requests
		2. Waits for in-flight requests
		3. Closes idle connections
		4. Respects the timeout context
	*/
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
