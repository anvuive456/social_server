package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	WebRTC   WebRTCConfig   `json:"webrtc"`
	Redis    RedisConfig    `json:"redis"`
}

type ServerConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	Environment     string        `json:"environment"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	MaxRequestSize  int64         `json:"max_request_size"`
	EnableCORS      bool          `json:"enable_cors"`
	AllowedOrigins  []string      `json:"allowed_origins"`
}

type DatabaseConfig struct {
	Host               string        `json:"host"`
	Port               string        `json:"port"`
	User               string        `json:"user"`
	Password           string        `json:"password"`
	Name               string        `json:"name"`
	SSLMode            string        `json:"ssl_mode"`
	ConnectTimeout     time.Duration `json:"connect_timeout"`
	QueryTimeout       time.Duration `json:"query_timeout"`
	MaxOpenConnections int           `json:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `json:"conn_max_lifetime"`
}

type AuthConfig struct {
	JWTSecret           string        `json:"jwt_secret"`
	JWTExpiration       time.Duration `json:"jwt_expiration"`
	RefreshExpiration   time.Duration `json:"refresh_expiration"`
	RSAPrivateKeyPath   string        `json:"rsa_private_key_path"`
	RSAPublicKeyPath    string        `json:"rsa_public_key_path"`
	PasswordMinLength   int           `json:"password_min_length"`
	MaxLoginAttempts    int           `json:"max_login_attempts"`
	LockoutDuration     time.Duration `json:"lockout_duration"`
	EnableTwoFactor     bool          `json:"enable_two_factor"`
}

type WebRTCConfig struct {
	STUNServers []string `json:"stun_servers"`
	TURNServers []struct {
		URL        string `json:"url"`
		Username   string `json:"username"`
		Credential string `json:"credential"`
	} `json:"turn_servers"`
	SignalingPort     string        `json:"signaling_port"`
	MaxRoomSize       int           `json:"max_room_size"`
	CallTimeout       time.Duration `json:"call_timeout"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

type RedisConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	Password        string        `json:"password"`
	Database        int           `json:"database"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	PoolSize        int           `json:"pool_size"`
	MinIdleConns    int           `json:"min_idle_conns"`
	MaxConnAge      time.Duration `json:"max_conn_age"`
	PoolTimeout     time.Duration `json:"pool_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	IdleCheckFreq   time.Duration `json:"idle_check_freq"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "localhost"),
			Port:            getEnv("SERVER_PORT", "8080"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxRequestSize:  getInt64Env("MAX_REQUEST_SIZE", 32<<20), // 32MB
			EnableCORS:      getBoolEnv("ENABLE_CORS", true),
			AllowedOrigins:  getStringSliceEnv("ALLOWED_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			Host:               getEnv("POSTGRES_HOST", "localhost"),
			Port:               getEnv("POSTGRES_PORT", "5432"),
			User:               getEnv("POSTGRES_USER", "postgres"),
			Password:           getEnv("POSTGRES_PASSWORD", ""),
			Name:               getEnv("POSTGRES_DATABASE", "social_media"),
			SSLMode:            getEnv("POSTGRES_SSLMODE", "disable"),
			ConnectTimeout:     getDurationEnv("POSTGRES_CONNECT_TIMEOUT", 10*time.Second),
			QueryTimeout:       getDurationEnv("POSTGRES_QUERY_TIMEOUT", 30*time.Second),
			MaxOpenConnections: getIntEnv("POSTGRES_MAX_OPEN_CONNECTIONS", 25),
			MaxIdleConnections: getIntEnv("POSTGRES_MAX_IDLE_CONNECTIONS", 5),
			ConnMaxLifetime:    getDurationEnv("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Auth: AuthConfig{
			JWTSecret:           getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			JWTExpiration:       getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			RefreshExpiration:   getDurationEnv("REFRESH_EXPIRATION", 7*24*time.Hour),
			RSAPrivateKeyPath:   getEnv("RSA_PRIVATE_KEY_PATH", "keys/private.pem"),
			RSAPublicKeyPath:    getEnv("RSA_PUBLIC_KEY_PATH", "keys/public.pem"),
			PasswordMinLength:   getIntEnv("PASSWORD_MIN_LENGTH", 8),
			MaxLoginAttempts:    getIntEnv("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:     getDurationEnv("LOCKOUT_DURATION", 15*time.Minute),
			EnableTwoFactor:     getBoolEnv("ENABLE_TWO_FACTOR", false),
		},
		WebRTC: WebRTCConfig{
			STUNServers: getStringSliceEnv("STUN_SERVERS", []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
			}),
			SignalingPort:     getEnv("WEBRTC_SIGNALING_PORT", "8081"),
			MaxRoomSize:       getIntEnv("WEBRTC_MAX_ROOM_SIZE", 10),
			CallTimeout:       getDurationEnv("WEBRTC_CALL_TIMEOUT", 30*time.Second),
			HeartbeatInterval: getDurationEnv("WEBRTC_HEARTBEAT_INTERVAL", 30*time.Second),
		},
		Redis: RedisConfig{
			Host:            getEnv("REDIS_HOST", "localhost"),
			Port:            getEnv("REDIS_PORT", "6379"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			Database:        getIntEnv("REDIS_DATABASE", 0),
			MaxRetries:      getIntEnv("REDIS_MAX_RETRIES", 3),
			RetryDelay:      getDurationEnv("REDIS_RETRY_DELAY", 200*time.Millisecond),
			PoolSize:        getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
			MaxConnAge:      getDurationEnv("REDIS_MAX_CONN_AGE", 30*time.Minute),
			PoolTimeout:     getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:     getDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
			IdleCheckFreq:   getDurationEnv("REDIS_IDLE_CHECK_FREQ", time.Minute),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if c.Auth.PasswordMinLength < 6 {
		return fmt.Errorf("password minimum length must be at least 6")
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func (c *Config) GetWebRTCAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.WebRTC.SignalingPort)
}

func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getUint64Env(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return []string{value}
	}
	return defaultValue
}