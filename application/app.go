package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/superlogarifm/goCalc-v3/internal/auth"
	"github.com/superlogarifm/goCalc-v3/internal/calculator"
	"github.com/superlogarifm/goCalc-v3/internal/http/handlers"
	"github.com/superlogarifm/goCalc-v3/internal/http/middleware"
	"github.com/superlogarifm/goCalc-v3/internal/storage"
	postgresrepo "github.com/superlogarifm/goCalc-v3/internal/storage/postgres"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host          string
	Port          string
	DatabaseURL   string // Строка подключения к БД
	JWTSecretKey  string // Секретный ключ для JWT
	TokenDuration time.Duration
}

func loadConfig() Config {
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("Warning: DATABASE_URL environment variable not set. Using default (likely invalid).")
		dbURL = "postgres://postgres:postgres@localhost:5432/gocalc?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Println("Warning: JWT_SECRET_KEY environment variable not set. Using default (insecure).")
		jwtSecret = "a-very-insecure-secret-key-replace-me"
	}
	tokeDurationStr := os.Getenv("TOKEN_DURATION")
	tokenDuration, err := time.ParseDuration(tokeDurationStr)
	if err != nil || tokenDuration <= 0 {
		log.Printf("Warning: Invalid or missing TOKEN_DURATION. Using default %v.\n", 24*time.Hour)
		tokenDuration = 24 * time.Hour // По умолчанию 24 часа
	}

	return Config{
		Host:          host,
		Port:          port,
		DatabaseURL:   dbURL,
		JWTSecretKey:  jwtSecret,
		TokenDuration: tokenDuration,
	}
}

type App struct {
	config           Config
	db               *gorm.DB
	authService      *auth.AuthService
	userRepo         storage.UserRepository
	taskManager      *calculator.TaskManager
	authHandlers     *handlers.AuthHandlers
	calculateHandler *handlers.CalculateHandler
	authMiddleware   *middleware.AuthMiddleware
	httpServer       *http.Server
}

func NewApp() *App {
	config := loadConfig()

	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established.")
	userRepo := postgresrepo.NewPGUserRepository(db)
	log.Println("Running database migrations...")
	err = userRepo.AutoMigrate()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed.")

	authService, err := auth.NewAuthService(config.JWTSecretKey, config.TokenDuration)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	taskManager := calculator.NewTaskManager()
	taskManager.StartInternalWorker()

	authHandlers := handlers.NewAuthHandlers(authService, userRepo)
	calculateHandler := handlers.NewCalculateHandler(taskManager)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	return &App{
		config:           config,
		db:               db,
		authService:      authService,
		userRepo:         userRepo,
		taskManager:      taskManager,
		authHandlers:     authHandlers,
		calculateHandler: calculateHandler,
		authMiddleware:   authMiddleware,
	}
}

func (a *App) StartServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", a.authHandlers.Register)
	mux.HandleFunc("/api/v1/login", a.authHandlers.Login)

	calculateMux := http.NewServeMux()
	calculateMux.HandleFunc("/api/v1/calculate", a.calculateHandler.HandleCalculate)
	calculateMux.HandleFunc("/api/v1/expressions", a.calculateHandler.HandleGetExpressions)     // Маршрут для GET /api/v1/expressions
	calculateMux.HandleFunc("/api/v1/expressions/", a.calculateHandler.HandleGetExpressionByID) // Маршрут для GET /api/v1/expressions/{id}

	protectedHandler := a.authMiddleware.Authenticate(calculateMux)
	mux.Handle("/api/v1/calculate", protectedHandler)
	mux.Handle("/api/v1/expressions", protectedHandler)
	mux.Handle("/api/v1/expressions/", protectedHandler)

	serverAddr := a.config.Host + ":" + a.config.Port
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	log.Printf("Starting server on %s\n", serverAddr)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", serverAddr, err)
		}
	}()
	log.Println("Server started.")
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	sqlDB, err := a.db.DB()
	if err == nil {
		log.Println("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v\n", err)
		} else {
			log.Println("Database connection closed.")
		}
	} else {
		log.Printf("Error getting underlying DB connection for closing: %v\n", err)
	}

	if a.httpServer != nil {
		return a.httpServer.Shutdown(ctx)
	}
	return nil
}
