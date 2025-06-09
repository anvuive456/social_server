package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"social_server/internal/config"
	"social_server/internal/database"
	"social_server/internal/middleware"
	"social_server/internal/repositories"
	"social_server/internal/repositories/postgres"
	"social_server/internal/routes"
	"social_server/internal/services"
	"syscall"
	"time"

	_ "social_server/docs"
)

// @title Social Media Backend API
// @version 2.0
// @description A comprehensive social media backend with posts, video calling, and file uploads
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 14.224.203.223:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Authentication
// @tag.description User authentication and authorization

// @tag.name Users
// @tag.description User management and profiles

// @tag.name Posts
// @tag.description Post creation, management and social interactions

// @tag.name Calls
// @tag.description WebRTC video calling functionality

// @tag.name Uploads
// @tag.description File upload and management

// @tag.name Search
// @tag.description Advanced search with Bleve full-text search

// @tag.name Chat
// @tag.description Real-time chat and messaging functionality

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup context with timeout for startup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize PostgreSQL
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error disconnecting from PostgreSQL: %v", err)
		}
	}()

	// Run auto-migration
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Create additional indexes
	if err := db.CreateIndexes(); err != nil {
		log.Printf("Warning: Failed to create some indexes: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	repos := &repositories.Repositories{
		User:             postgres.NewUserRepository(db.DB),
		Friend:           postgres.NewFriendRepository(db.DB),
		Profile:          postgres.NewProfileRepository(db.DB),
		Post:             postgres.NewPostRepository(db.DB),
		Comment:          postgres.NewCommentRepository(db.DB),
		Like:             postgres.NewLikeRepository(db.DB),
		Share:            postgres.NewShareRepository(db.DB),
		ChatRoom:         postgres.NewChatRoomRepository(db.DB),
		Message:          postgres.NewMessageRepository(db.DB),
		Participant:      postgres.NewParticipantRepository(db.DB),
		TypingIndicator:  postgres.NewTypingIndicatorRepository(db.DB),
		OnlineStatus:     postgres.NewOnlineStatusRepository(db.DB),
		ChatInvite:       postgres.NewChatInviteRepository(db.DB),
		ChatNotification: postgres.NewChatNotificationRepository(db.DB),
		Auth:             postgres.NewAuthRepository(db.DB),
		Call:             postgres.NewCallRepository(db.DB),
		Session:          postgres.NewSessionRepository(db.DB),
	}

	// Initialize services
	authService, err := services.NewAuthService(repos.User, repos.Auth, repos.Profile, repos.Session, &cfg.Auth)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	postService := services.NewPostService(
		repos.Post,
		repos.User,
		repos.Comment,
		repos.Like,
		repos.Share,
	)

	callService := services.NewCallService(
		repos.Call,
		repos.User,
	)

	searchService, err := services.NewSearchService("./search_index", repos.Post, repos.User)
	if err != nil {
		log.Fatalf("Failed to initialize search service: %v", err)
	}

	chatService := services.NewChatService(repos)

	userService := services.NewUserService(repos.User, &cfg.Auth)

	friendService := services.NewFriendService(repos.User, repos.Friend)

	profileService := services.NewProfileService(repos.Profile, repos.User, &cfg.Auth)

	mailService := services.NewMailService()

	// Initialize online status service
	onlineStatusService := services.NewOnlineStatusService(db.DB)

	// Configure session security service
	securityConfig := &services.SecurityConfig{
		MaxFailedAttempts:   5,
		LockoutDuration:     15 * time.Minute,
		SessionTimeout:      24 * time.Hour,
		MaxSessionsPerUser:  5,
		SuspiciousThreshold: 3,
	}
	sessionSecService := services.NewSessionSecurityService(repos.Session, repos.User, securityConfig)

	log.Println("Services initialized successfully")

	// Configure middleware factory
	middlewareConfig := &middleware.MiddlewareConfig{
		MaxSessionsPerUser:  5,
		RateLimitPerMinute:  60,
		SessionTimeout:      24 * time.Hour,
		MaxFailedAttempts:   5,
		LockoutDuration:     15 * time.Minute,
		SuspiciousThreshold: 3,
	}

	// Initialize global middleware factory for convenience functions
	middleware.InitializeMiddlewareFactory(
		authService,
		sessionSecService,
		repos.Session,
		repos.User,
		middlewareConfig,
	)

	// Setup routes
	router := routes.NewRouter(
		cfg,
		authService,
		userService,
		profileService,
		friendService,
		postService,
		chatService,
		searchService,
		callService,
		mailService,
		onlineStatusService,
	)
	engine := router.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Social Media Server started successfully on %s", cfg.GetServerAddress())
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Health check available at: http://%s/health", cfg.GetServerAddress())
	log.Printf("API documentation available at: http://%s/swagger/index.html", cfg.GetServerAddress())

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Cleanup online status service
	onlineStatusService.Stop()

	log.Println("Server shutdown completed")
}
