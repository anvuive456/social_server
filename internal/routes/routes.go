package routes

import (
	"social_server/internal/config"
	"social_server/internal/handlers"
	"social_server/internal/middleware"
	"social_server/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	config              *config.Config
	authHandler         *handlers.AuthHandler
	userHandler         *handlers.UserHandler
	profileHandler      *handlers.ProfileHandler
	friendHandler       *handlers.FriendHandler
	postHandler         *handlers.PostHandler
	chatHandler         *handlers.ChatHandler
	uploadHandler       *handlers.UploadHandler
	searchHandler       *handlers.SearchHandler
	callHandler         *handlers.CallHandler
	wsHandler           *handlers.WebSocketHandler
	onlineStatusHandler *handlers.OnlineStatusHandler
}

func NewRouter(
	cfg *config.Config,
	authService *services.AuthService,
	userService *services.UserService,
	profileService *services.ProfileService,
	friendService *services.FriendService,
	postService *services.PostService,
	chatService *services.ChatService,
	searchService *services.SearchService,
	callService *services.CallService,
	mailService *services.MailService,
	onlineStatusService *services.OnlineStatusService,
) *Router {
	wsHandler := handlers.NewWebSocketHandler(authService, callService, chatService)

	onlineStatusHandler := handlers.NewOnlineStatusHandler(onlineStatusService, authService)

	return &Router{
		config:              cfg,
		authHandler:         handlers.NewAuthHandler(authService, mailService),
		userHandler:         handlers.NewUserHandler(userService),
		profileHandler:      handlers.NewProfileHandler(profileService),
		friendHandler:       handlers.NewFriendHandler(friendService),
		postHandler:         handlers.NewPostHandler(postService),
		chatHandler:         handlers.NewChatHandler(chatService, userService),
		uploadHandler:       handlers.NewUploadHandler("./uploads"),
		searchHandler:       handlers.NewSearchHandler(searchService),
		callHandler:         handlers.NewCallHandler(callService, wsHandler),
		wsHandler:           wsHandler,
		onlineStatusHandler: onlineStatusHandler,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	// Set Gin mode
	if r.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())

	// CORS middleware
	corsConfig := &middleware.CORSConfig{
		AllowOrigins:     r.config.Server.AllowedOrigins,
		AllowCredentials: true,
	}
	router.Use(middleware.CORS(corsConfig))

	// Logging middleware
	if r.config.IsDevelopment() {
		router.Use(middleware.Logging(nil))
	} else {
		router.Use(middleware.JSONLogging())
	}

	// Rate limiting middleware
	router.Use(middleware.RateLimitByIP(100, 20)) // 100 requests per minute, burst of 20

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "social_server",
			"version": "2.0.0",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static file serving for uploads
	router.Static("/uploads", "./uploads")

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		r.setupAuthRoutes(v1)
		r.setupUserRoutes(v1)
		r.setupProfileRoutes(v1)
		r.setupFriendRoutes(v1)
		r.setupPostRoutes(v1)
		r.setupChatRoutes(v1)
		r.setupSearchRoutes(v1)
		r.setupCallRoutes(v1)
		r.setupOnlineStatusRoutes(v1)
	}

	return router
}

func (r *Router) setupAuthRoutes(v1 *gin.RouterGroup) {
	auth := v1.Group("/auth")
	{
		// Higher rate limiting for auth endpoints
		auth.Use(middleware.RateLimitAuth(10, 5)) // 10 requests per minute, burst of 5

		// Public auth routes
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)
		auth.POST("/forgot-password", r.authHandler.ForgotPassword)
		auth.POST("/reset-password", r.authHandler.ResetPassword)
		auth.GET("/verify-email", r.authHandler.VerifyEmail)
		auth.POST("/resend-verification", r.authHandler.ResendVerification)

		// Protected auth routes
		authProtected := auth.Group("")
		authProtected.Use(middleware.Auth())
		{
			authProtected.POST("/logout", r.authHandler.Logout)
			authProtected.POST("/change-password", r.authHandler.ChangePassword)
			// authProtected.GET("/sessions", r.authHandler.GetSessions)
			// authProtected.DELETE("/sessions/:session_id", r.authHandler.RevokeSession)
			// authProtected.POST("/sessions/revoke-all", r.authHandler.RevokeAllSessions)
		}
	}
}

func (r *Router) setupUserRoutes(v1 *gin.RouterGroup) {
	users := v1.Group("/users")
	users.Use(middleware.Auth())
	{
		// Public user routes
		users.GET("/search", r.userHandler.SearchUsers)
		users.GET("/:id/stats", r.userHandler.GetUserStats)

		// Online status and settings
		users.PUT("/online-status", r.userHandler.UpdateOnlineStatus)
		users.GET("/settings", r.userHandler.GetSettings)
		users.PUT("/settings", r.userHandler.UpdateSettings)

	}
}

func (r *Router) setupFriendRoutes(v1 *gin.RouterGroup) {
	friends := v1.Group("/friends")
	friends.Use(middleware.Auth())
	{
		// Friends list
		friends.GET("", r.friendHandler.GetFriends)
		// friends.GET("/user/:id", r.friendHandler.GetUserFriends)
		friends.GET("/blocked", r.friendHandler.GetBlockedUsers)

		// Friend requests
		friends.GET("/requests", r.friendHandler.GetFriendRequests)
		friends.GET("/requests-stats", r.friendHandler.GetFriendRequestStats)
		friends.POST("/send-request", r.friendHandler.SendFriendRequest)
		friends.POST("/accept-request", r.friendHandler.AcceptFriendRequest)
		friends.POST("/decline-request", r.friendHandler.DeclineFriendRequest)

		// Friend management
		friends.DELETE("/:id", r.friendHandler.RemoveFriend)
		friends.POST("/:id/block", r.friendHandler.BlockUser)
		friends.POST("/:id/unblock", r.friendHandler.UnblockUser)
		friends.GET("/:id/status", r.friendHandler.CheckFriendship)
	}
}

func (r *Router) setupPostRoutes(v1 *gin.RouterGroup) {
	posts := v1.Group("/posts")
	posts.Use(middleware.Auth())

	{
		// Posts
		posts.GET("", r.postHandler.GetPosts)
		posts.GET("/:id", r.postHandler.GetPost)

		// Post management
		posts.POST("", r.postHandler.CreatePost)
		posts.PUT("/:id", r.postHandler.UpdatePost)
		posts.DELETE("/:id", r.postHandler.DeletePost)

		// Feed
		posts.GET("/feed", r.postHandler.GetFeed)

		// Post interactions
		posts.POST("/:id/like", r.postHandler.LikePost)
		posts.POST("/:id/view", r.postHandler.ViewPost)
		posts.POST("/:id/comments", r.postHandler.CreateComment)
		posts.POST("/:id/share", r.postHandler.SharePost)
	}

}

func (r *Router) setupChatRoutes(v1 *gin.RouterGroup) {
	chat := v1.Group("/chat")
	chat.Use(middleware.Auth())
	{
		// Room management
		chat.POST("/rooms", r.chatHandler.CreateRoom)
		chat.GET("/rooms", r.chatHandler.GetRooms)
		// chat.GET("/rooms/search", r.chatHandler.SearchRooms)
		// chat.GET("/rooms/:room_id", r.chatHandler.GetRoom)
		// // chat.PUT("/rooms/:room_id", r.chatHandler.UpdateRoom)
		chat.DELETE("/rooms/:id", r.chatHandler.DeleteRoom)
		chat.GET("/rooms/sync", r.chatHandler.SyncRooms)

		// // Message management
		// chat.GET("/rooms/:room_id/messages", r.chatHandler.GetMessages)
		// chat.POST("/rooms/:room_id/messages", r.chatHandler.SendMessage)
		// // chat.PUT("/messages/:message_id", r.chatHandler.UpdateMessage)
		// chat.DELETE("/messages/:message_id", r.chatHandler.DeleteMessage)
		// chat.POST("/messages/:message_id/read", r.chatHandler.MarkMessageRead)

		// // Participant management
		// chat.GET("/rooms/:room_id/participants", r.chatHandler.GetParticipants)
		// chat.POST("/rooms/:room_id/participants", r.chatHandler.AddParticipant)
		// chat.DELETE("/rooms/:room_id/participants/:user_id", r.chatHandler.RemoveParticipant)
		// // chat.PUT("/rooms/:room_id/participants/:user_id/role", r.chatHandler.UpdateParticipantRole)

		// // Reactions and typing
		// chat.POST("/messages/:message_id/reactions", r.chatHandler.AddReaction)
		// chat.DELETE("/messages/:message_id/reactions/:emoji", r.chatHandler.RemoveReaction)
		// chat.POST("/rooms/:room_id/typing", r.chatHandler.SetTyping)
	}
}

func (r *Router) setupSearchRoutes(v1 *gin.RouterGroup) {
	search := v1.Group("/search")
	{
		// Public search endpoints
		search.GET("", r.searchHandler.Search)
		search.GET("/posts", r.searchHandler.SearchPosts)
		search.GET("/users", r.searchHandler.SearchUsers)
		search.GET("/autocomplete", r.searchHandler.AutoComplete)
		search.GET("/trending-tags", r.searchHandler.GetTrendingTags)

	}
}

func (r *Router) setupCallRoutes(v1 *gin.RouterGroup) {
	calls := v1.Group("/calls")
	calls.Use(middleware.Auth())
	{
		// Call management
		calls.POST("", r.callHandler.InitiateCall)
		calls.GET("/history", r.callHandler.GetCallHistory)
		calls.GET("/active", r.callHandler.GetActiveCalls)
		calls.GET("/stats", r.callHandler.GetCallStats)
		calls.GET("/webrtc-config", r.callHandler.GetWebRTCConfig)
		calls.GET("/connection-stats", r.callHandler.GetConnectionStats)
		calls.GET("/:id", r.callHandler.GetCall)
		calls.POST("/:id/accept", r.callHandler.AcceptCall)
		calls.POST("/:id/decline", r.callHandler.DeclineCall)
		calls.POST("/:id/end", r.callHandler.EndCall)
	}

	// WebSocket routes for signaling
	callGroup := v1.Group("/ws")
	callGroup.Use(middleware.Auth())
	{
		callGroup.GET("/calls", r.wsHandler.HandleWebSocket)
	}

	// v1.GET("/ws/chat", middleware.Auth(), r.chatHandler.HandleWebSocket)
}

func (r *Router) setupProfileRoutes(v1 *gin.RouterGroup) {
	// Profile routes for current user (1-1 relationship)
	profile := v1.Group("/profile")
	profile.Use(middleware.Auth())
	{
		// Single profile management for current user
		profile.GET("", r.profileHandler.GetMyProfile)
		profile.POST("", r.profileHandler.CreateOrUpdateProfile)
		profile.PUT("", r.profileHandler.UpdateProfile)
		// profile.DELETE("", r.profileHandler.DeleteProfile)
	}
}

func (r *Router) setupOnlineStatusRoutes(v1 *gin.RouterGroup) {
	onlineStatus := v1.Group("/online-status")
	onlineStatus.Use(middleware.Auth())
	{
		// User online status management
		onlineStatus.GET("/me", r.onlineStatusHandler.GetMyOnlineStatus)
		onlineStatus.GET("/user/:user_id", r.onlineStatusHandler.GetUserOnlineStatus)
		onlineStatus.GET("/user/:user_id/last-seen", r.onlineStatusHandler.GetUserLastSeen)

		// Friends and social features
		onlineStatus.GET("/friends", r.onlineStatusHandler.GetOnlineFriends)
		onlineStatus.GET("/count", r.onlineStatusHandler.GetOnlineUsersCount)

		// Admin/monitoring endpoints
		onlineStatus.GET("/users", r.onlineStatusHandler.GetOnlineUsers)
		onlineStatus.GET("/stats", r.onlineStatusHandler.GetConnectionStats)
		onlineStatus.POST("/user/:user_id/offline", r.onlineStatusHandler.SetUserOffline)
	}
}
