package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"social_server/internal/config"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repositories.UserRepository
	authRepo    repositories.AuthRepository
	profileRepo repositories.ProfileRepository
	sessionRepo repositories.SessionRepository
	config      *config.AuthConfig
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	TokenID   string `json:"token_id"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	TokenID   string `json:"token_id"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo repositories.UserRepository, authRepo repositories.AuthRepository, profileRepo repositories.ProfileRepository, sessionRepo repositories.SessionRepository, config *config.AuthConfig) (*AuthService, error) {
	service := &AuthService{
		userRepo:    userRepo,
		authRepo:    authRepo,
		profileRepo: profileRepo,
		sessionRepo: sessionRepo,
		config:      config,
	}

	// Load or generate RSA keys
	if err := service.loadOrGenerateKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize RSA keys: %w", err)
	}

	return service, nil
}

func (s *AuthService) loadOrGenerateKeys() error {
	privateKeyPath := s.config.RSAPrivateKeyPath
	publicKeyPath := s.config.RSAPublicKeyPath

	// Try to load existing keys
	if _, err := os.Stat(privateKeyPath); err == nil {
		return s.loadKeys(privateKeyPath, publicKeyPath)
	}

	// Generate new keys if they don't exist
	return s.generateAndSaveKeys(privateKeyPath, publicKeyPath)
}

func (s *AuthService) loadKeys(privateKeyPath, publicKeyPath string) error {
	// Load private key
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA format")
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	return nil
}

func (s *AuthService) generateAndSaveKeys(privateKeyPath, publicKeyPath string) error {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	// Save private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

func (s *AuthService) Register(req *requests.RegisterRequest, ipAddress, userAgent string) (*responses.AuthResponse, error) {
	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &postgres.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		IsVerified:   false,
		Settings: postgres.UserSettings{
			PrivacyProfileVisibility:    "public",
			PrivacyShowOnlineStatus:     true,
			PrivacyAllowFriendRequests:  true,
			NotificationsEmail:          true,
			NotificationsPush:           true,
			NotificationsFriendRequests: true,
			NotificationsMessages:       true,
			NotificationsPosts:          true,
		},
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Implement email verification creation
	// verificationToken := uuid.New().String()
	// emailVerification := &postgres.EmailVerification{
	// 	UserID:    user.ID,
	// 	Token:     verificationToken,
	// 	Email:     user.Email,
	// 	Type:      "registration",
	// 	ExpiresAt: time.Now().Add(24 * time.Hour),
	// 	IPAddress: ipAddress,
	// 	UserAgent: userAgent,
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }
	// err = s.authRepo.CreateEmailVerification( emailVerification)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create email verification: %w", err)
	// }

	// Log security event
	s.logSecurityEvent(&user.ID, "user_registration", "New user registered", ipAddress, userAgent, "info")

	// Generate tokens
	sessionID := uuid.New().String()
	tokenPair, tokenID, err := s.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &postgres.Session{
		UserID:       user.ID,
		TokenID:      tokenID,
		SessionID:    sessionID,
		RefreshToken: tokenPair.RefreshToken,
		IsActive:     true,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(s.config.RefreshExpiration),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = s.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &responses.AuthResponse{
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		ExpiresIn:        int64(s.config.JWTExpiration.Seconds()),
		RefreshExpiresIn: int64(s.config.RefreshExpiration.Seconds()),
		TokenType:        "Bearer",
		User: responses.UserAuth{
			ID:       createdUser.ID,
			Username: createdUser.Email,
			Email:    createdUser.Email,
			IsActive: createdUser.IsActive,
		},
	}, nil
}

func (s *AuthService) Login(req *requests.LoginRequest, ipAddress, userAgent string) (*responses.AuthResponse, error) {
	// Log login attempt
	s.logLoginAttempt(req.EmailOrUsername, ipAddress, userAgent, false, "")

	// Rate limiting check
	if err := s.checkRateLimit(ipAddress, "login"); err != nil {
		return nil, err
	}

	// Find user by email or username
	var user *postgres.User
	var err error

	if strings.Contains(req.EmailOrUsername, "@") {
		user, err = s.userRepo.GetByEmail(req.EmailOrUsername)
	} else {
		user, err = s.userRepo.GetByUsername(req.EmailOrUsername)
	}

	if err != nil {
		s.logLoginAttempt(req.EmailOrUsername, ipAddress, userAgent, false, "user_not_found")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is active
	if !user.IsActive {
		s.logLoginAttempt(req.EmailOrUsername, ipAddress, userAgent, false, "account_inactive")
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		s.logLoginAttempt(req.EmailOrUsername, ipAddress, userAgent, false, "invalid_password")
		s.logSecurityEvent(&user.ID, "failed_login", "Failed login attempt", ipAddress, userAgent, "warning")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update online status
	err = s.userRepo.Update(user.ID, map[string]interface{}{
		"is_online":  true,
		"last_seen":  time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Generate session and tokens
	sessionID := uuid.New().String()
	tokenPair, tokenID, err := s.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create login session
	sessionExpiry := time.Now().Add(s.config.RefreshExpiration)
	if req.RememberMe {
		sessionExpiry = time.Now().Add(30 * 24 * time.Hour) // 30 days
	}

	// Create session
	session := &postgres.Session{
		UserID:       user.ID,
		TokenID:      tokenID,
		SessionID:    sessionID,
		RefreshToken: tokenPair.RefreshToken,
		IsActive:     true,
		DeviceInfo:   req.DeviceInfo,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		Platform:     req.Platform,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    sessionExpiry,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = s.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful login
	s.logLoginAttempt(req.EmailOrUsername, ipAddress, userAgent, true, "")
	s.logSecurityEvent(&user.ID, "successful_login", "User logged in successfully", ipAddress, userAgent, "info")

	return &responses.AuthResponse{
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		ExpiresIn:        int64(s.config.JWTExpiration.Seconds()),
		RefreshExpiresIn: int64(s.config.RefreshExpiration.Seconds()),
		TokenType:        "Bearer",
		User: responses.UserAuth{
			ID:       user.ID,
			Username: user.Email,
			Email:    user.Email,
			IsActive: user.IsActive,
		},
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string, ipAddress, userAgent string) (*responses.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logSecurityEvent(nil, "invalid_refresh_token", "Invalid refresh token used", ipAddress, userAgent, "warning")
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if user exists and is active
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Validate session
	session, err := s.sessionRepo.GetSessionBySessionID(claims.SessionID)
	if err != nil || !session.IsValidActive() {
		s.logSecurityEvent(&user.ID, "invalid_session", "Session expired or invalid during refresh", ipAddress, userAgent, "warning")
		return nil, fmt.Errorf("session expired or invalid")
	}

	// Generate new token pair
	newSessionID := uuid.New().String()
	tokenPair, tokenID, err := s.GenerateTokenPair(user.ID, newSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Update session with new token info
	session.TokenID = tokenID
	session.SessionID = newSessionID
	session.RefreshToken = tokenPair.RefreshToken
	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()
	
	err = s.sessionRepo.UpdateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &responses.AuthResponse{
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		ExpiresIn:        int64(s.config.JWTExpiration.Seconds()),
		RefreshExpiresIn: int64(s.config.RefreshExpiration.Seconds()),
		TokenType:        "Bearer",
		User: responses.UserAuth{
			ID:       user.ID,
			Username: user.Email,
			Email:    user.Email,
			IsActive: user.IsActive,
		},
	}, nil
}

func (s *AuthService) Logout(userID uint, sessionID string) error {
	// TODO: Implement session deactivation
	// err := s.authRepo.DeactivateSession( sessionID)
	// if err != nil {
	// 	return fmt.Errorf("failed to deactivate session: %w", err)
	// }

	// Update user online status
	err := s.userRepo.Update(userID, map[string]interface{}{
		"is_online":  false,
		"last_seen":  time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (s *AuthService) GenerateTokenPair(userID uint, sessionID string) (*TokenPair, string, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, "", fmt.Errorf("user not found: %w", err)
	}

	tokenID := uuid.New().String()
	now := time.Now()

	// Generate access token
	accessClaims := &Claims{
		UserID:    userID,
		Username:  user.Email,
		Email:     user.Email,
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "social_server",
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{"social_app"},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.JWTExpiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "social_server",
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{"social_app"},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExpiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token
	refreshTokenEntity := &postgres.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: now.Add(s.config.RefreshExpiration),
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = s.authRepo.CreateRefreshToken(refreshTokenEntity)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
		TokenType:    "Bearer",
	}, tokenID, nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token claims")
}

func (s *AuthService) ChangePassword(userID uint, req *requests.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	updates := map[string]interface{}{
		"password_hash": string(hashedPassword),
		"updated_at":    time.Now(),
	}

	err = s.userRepo.Update(userID, updates)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// TODO: Implement session revocation
	// err = s.authRepo.RevokeUserSessions( userID)
	// if err != nil {
	// 	return fmt.Errorf("failed to revoke sessions: %w", err)
	// }

	return nil
}

func (s *AuthService) ForgotPassword(req *requests.ForgotPasswordRequest, ipAddress, userAgent string) error {
	_, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// TODO: Implement password reset creation
	// resetToken := uuid.New().String()
	// passwordReset := &postgres.PasswordReset{
	// 	UserID:    user.ID,
	// 	Token:     resetToken,
	// 	Email:     req.Email,
	// 	ExpiresAt: time.Now().Add(time.Hour), // 1 hour expiry
	// 	IPAddress: ipAddress,
	// 	UserAgent: userAgent,
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }
	// err = s.authRepo.CreatePasswordReset( passwordReset)
	// if err != nil {
	// 	return fmt.Errorf("failed to create password reset: %w", err)
	// }

	// TODO: Send email with reset token

	return nil
}

func (s *AuthService) ResetPassword(req *requests.ResetPasswordRequest, ipAddress, userAgent string) error {
	// TODO: Implement password reset validation
	// passwordReset, err := s.authRepo.GetPasswordResetByToken( req.Token)
	// if err != nil {
	// 	return fmt.Errorf("invalid or expired reset token")
	// }
	return fmt.Errorf("password reset not implemented yet")
}

func (s *AuthService) VerifyEmail(req *requests.VerifyEmailRequest) error {
	// TODO: Implement email verification
	// verification, err := s.authRepo.GetEmailVerificationByToken( req.Token)
	// if err != nil {
	// 	return fmt.Errorf("invalid or expired verification token")
	// }
	return fmt.Errorf("email verification not implemented yet")
}

// Helper methods
func (s *AuthService) validateRegisterRequest(req *requests.RegisterRequest) error {

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if !req.AcceptTerms || !req.AcceptPrivacy {
		return fmt.Errorf("must accept terms and privacy policy")
	}

	return nil
}

func (s *AuthService) checkRateLimit(key, action string) error {
	// TODO: Implement rate limiting logic
	return nil
}

func (s *AuthService) logLoginAttempt(email, ipAddress, userAgent string, successful bool, failureReason string) {
	attempt := &postgres.LoginAttempt{
		Email:         email,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		IsSuccessful:  successful,
		FailureReason: failureReason,
		AttemptedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}
	s.authRepo.LogLoginAttempt(attempt)
}

func (s *AuthService) logSecurityEvent(userID *uint, eventType, description, ipAddress, userAgent, severity string) {
	event := &postgres.SecurityEvent{
		UserID:      userID,
		EventType:   eventType,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    severity,
		CreatedAt:   time.Now(),
	}
	s.authRepo.CreateSecurityEvent(event)
}
