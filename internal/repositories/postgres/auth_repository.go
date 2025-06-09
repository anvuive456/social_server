package postgres

import (
	"context"
	"time"

	"social_server/internal/models/postgres"
	"social_server/internal/repositories"

	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) repositories.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateRefreshToken(token *postgres.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *authRepository) GetRefreshToken(token string) (*postgres.RefreshToken, error) {
	var refreshToken postgres.RefreshToken
	err := r.db.
		Where("token = ? AND is_revoked = ? AND expires_at > ?", token, false, time.Now()).
		Preload("User").
		First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *authRepository) RevokeRefreshToken(token string) error {
	return r.db.
		Model(&postgres.RefreshToken{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"updated_at": time.Now(),
		}).Error
}

func (r *authRepository) CleanupExpiredTokens(ctx context.Context) error {
	return r.db.
		Where("expires_at < ? OR is_revoked = ?", time.Now(), true).
		Delete(&postgres.RefreshToken{}).Error
}

func (r *authRepository) CreateLoginSession(session *postgres.LoginSession) error {
	return r.db.Create(session).Error
}

func (r *authRepository) GetLoginSession(sessionID string) (*postgres.LoginSession, error) {
	var session postgres.LoginSession
	err := r.db.
		Where("session_id = ? AND is_active = ? AND expires_at > ?", sessionID, true, time.Now()).
		Preload("User").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *authRepository) UpdateLoginSession(sessionID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.LoginSession{}).
		Where("session_id = ?", sessionID).
		Updates(updates).Error
}

func (r *authRepository) DeleteLoginSession(sessionID string) error {
	return r.db.
		Where("session_id = ?", sessionID).
		Delete(&postgres.LoginSession{}).Error
}

func (r *authRepository) CreatePasswordReset(reset *postgres.PasswordReset) error {
	return r.db.Create(reset).Error
}

func (r *authRepository) GetPasswordReset(token string) (*postgres.PasswordReset, error) {
	var reset postgres.PasswordReset
	err := r.db.
		Where("token = ? AND is_used = ? AND expires_at > ?", token, false, time.Now()).
		Preload("User").
		First(&reset).Error
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

func (r *authRepository) UsePasswordReset(token string) error {
	now := time.Now()
	return r.db.
		Model(&postgres.PasswordReset{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_used":    true,
			"used_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *authRepository) CreateEmailVerification(verification *postgres.EmailVerification) error {
	return r.db.Create(verification).Error
}

func (r *authRepository) GetEmailVerification(token string) (*postgres.EmailVerification, error) {
	var verification postgres.EmailVerification
	err := r.db.
		Where("token = ? AND is_verified = ? AND expires_at > ?", token, false, time.Now()).
		Preload("User").
		First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *authRepository) VerifyEmail(token string) error {
	now := time.Now()
	return r.db.
		Model(&postgres.EmailVerification{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_verified": true,
			"verified_at": &now,
			"updated_at":  now,
		}).Error
}

func (r *authRepository) LogLoginAttempt(attempt *postgres.LoginAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *authRepository) GetLoginAttempts(email, ipAddress string, since time.Time) ([]postgres.LoginAttempt, error) {
	var attempts []postgres.LoginAttempt
	db := r.db.
		Where("attempted_at > ?", since)

	if email != "" {
		db = db.Where("email = ?", email)
	}
	if ipAddress != "" {
		db = db.Where("ip_address = ?", ipAddress)
	}

	err := db.Order("attempted_at DESC").Find(&attempts).Error
	return attempts, err
}

func (r *authRepository) CreateSecurityEvent(event *postgres.SecurityEvent) error {
	return r.db.Create(event).Error
}

func (r *authRepository) CheckRateLimit(key string, windowStart time.Time, limit int) (bool, error) {
	var entry postgres.RateLimitEntry
	err := r.db.
		Where("key = ? AND window_start = ?", key, windowStart).
		First(&entry).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil // No entry found, rate limit not exceeded
		}
		return false, err
	}

	return entry.Count < limit, nil
}

func (r *authRepository) IncrementRateLimit(key string, windowStart, expiresAt time.Time) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		var entry postgres.RateLimitEntry
		err := tx.Where("key = ? AND window_start = ?", key, windowStart).First(&entry).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new entry
				entry = postgres.RateLimitEntry{
					Key:         key,
					Count:       1,
					WindowStart: windowStart,
					ExpiresAt:   expiresAt,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				return tx.Create(&entry).Error
			}
			return err
		}

		// Update existing entry
		return tx.Model(&entry).Updates(map[string]interface{}{
			"count":      gorm.Expr("count + 1"),
			"updated_at": now,
		}).Error
	})
}
