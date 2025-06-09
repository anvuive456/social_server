package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"social_server/internal/config"
	models "social_server/internal/models/postgres"
)

type PostgresDB struct {
	*gorm.DB
}

func NewPostgresConnection(cfg *config.Config) (*PostgresDB, error) {
	// Setup GORM logger
	var gormLogger logger.Interface
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Create GORM config
	gormConfig := &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.GetPostgresDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying sql.DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL at %s:%s", cfg.Database.Host, cfg.Database.Port)

	return &PostgresDB{DB: db}, nil
}

func (pg *PostgresDB) AutoMigrate() error {
	// Auto-migrate all models
	err := pg.DB.AutoMigrate(
		// User models
		&models.User{},
		&models.Profile{},
		&models.FriendRequest{},
		&models.UserFriend{},

		// Auth models
		&models.RefreshToken{},
		&models.LoginSession{},
		&models.PasswordReset{},
		&models.EmailVerification{},
		&models.LoginAttempt{},
		&models.RateLimitEntry{},
		&models.SecurityEvent{},
		&models.TwoFactorAuth{},

		// Chat models
		&models.ChatRoom{},
		&models.Participant{},
		&models.Message{},
		&models.MessageRead{},
		&models.MessageReaction{},
		&models.TypingIndicator{},
		&models.OnlineStatus{},
		&models.ChatInvite{},
		&models.ChatNotification{},

		// Post models
		&models.Post{},
		&models.PostMedia{},
		&models.Like{},
		&models.Comment{},
		&models.CommentLike{},
		&models.Share{},
		&models.PostView{},
		&models.PostReport{},
		&models.SavedPost{},
		&models.PostTag{},

		// Session models
		&models.Session{},
		&models.TokenBlacklist{},
		&models.SecurityLog{},

		// Call models
		&models.Call{},
		&models.CallParticipant{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	log.Println("Successfully auto-migrated all models")
	return nil
}

func (pg *PostgresDB) CreateIndexes() error {
	// Create additional indexes for performance
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email) WHERE is_active = true",
		// "CREATE INDEX IF NOT EXISTS idx_users_username_active ON users(username) WHERE is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_users_online ON users(is_online) WHERE is_online = true",

		// User friend indexes
		"CREATE INDEX IF NOT EXISTS idx_users_friend_active ON users(id) WHERE is_active = true",

		// Friend request indexes
		"CREATE INDEX IF NOT EXISTS idx_friend_requests_sender_status ON friend_requests(sender_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver_status ON friend_requests(receiver_id, status)",

		// Chat room indexes
		"CREATE INDEX IF NOT EXISTS idx_chat_rooms_last_activity ON chat_rooms(last_activity DESC)",
		"CREATE INDEX IF NOT EXISTS idx_chat_rooms_type_archived ON chat_rooms(type) WHERE is_archived = false",

		// Message indexes
		"CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages(chat_room_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender_created ON messages(sender_id, created_at DESC)",

		// Post indexes
		"CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts(author_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_posts_privacy_created ON posts(privacy, created_at DESC) WHERE deleted_at IS NULL",

		// Comment indexes
		"CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL",

		// Like indexes
		"CREATE INDEX IF NOT EXISTS idx_likes_post_user ON likes(post_id, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_likes_user_created ON likes(user_id, created_at DESC)",

		// Login attempt indexes
		"CREATE INDEX IF NOT EXISTS idx_login_attempts_email_time ON login_attempts(email, attempted_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_time ON login_attempts(ip_address, attempted_at DESC)",

		// Rate limit indexes
		"CREATE INDEX IF NOT EXISTS idx_rate_limits_expires ON rate_limit_entries(expires_at)",

		// Online status indexes
		"CREATE INDEX IF NOT EXISTS idx_online_status_user ON online_statuses(user_id, is_online)",

		// Session indexes
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_token_id ON sessions(token_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_device ON sessions(device_id) WHERE device_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_sessions_ip ON sessions(ip_address)",

		// Token blacklist indexes
		"CREATE INDEX IF NOT EXISTS idx_token_blacklist_token ON token_blacklist(token_id)",
		"CREATE INDEX IF NOT EXISTS idx_token_blacklist_user ON token_blacklist(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires ON token_blacklist(expires_at)",

		// Security log indexes
		"CREATE INDEX IF NOT EXISTS idx_security_logs_user_created ON security_logs(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_security_logs_session ON security_logs(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_security_logs_event_type ON security_logs(event_type, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_security_logs_risk ON security_logs(risk) WHERE risk IN ('high', 'critical')",
		"CREATE INDEX IF NOT EXISTS idx_security_logs_ip ON security_logs(ip_address, created_at DESC)",

		// Call indexes
		"CREATE INDEX IF NOT EXISTS idx_calls_caller_created ON calls(caller_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_calls_callee_created ON calls(callee_id, created_at DESC) WHERE callee_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_calls_status ON calls(status, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_calls_type ON calls(type, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_calls_room_id ON calls(room_id) WHERE room_id IS NOT NULL",

		// Call participant indexes
		"CREATE INDEX IF NOT EXISTS idx_call_participants_call ON call_participants(call_id, joined_at)",
		"CREATE INDEX IF NOT EXISTS idx_call_participants_user ON call_participants(user_id, joined_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_call_participants_active ON call_participants(call_id, is_active) WHERE is_active = true",
	}

	for _, index := range indexes {
		if err := pg.DB.Exec(index).Error; err != nil {
			log.Printf("Warning: failed to create index: %s, error: %v", index, err)
		}
	}

	log.Println("Successfully created additional indexes")
	return nil
}

func (pg *PostgresDB) Close() error {
	sqlDB, err := pg.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("Database connection closed successfully")
	return nil
}

func (pg *PostgresDB) GetStats() map[string]interface{} {
	sqlDB, err := pg.DB.DB()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

func (pg *PostgresDB) HealthCheck() error {
	sqlDB, err := pg.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
