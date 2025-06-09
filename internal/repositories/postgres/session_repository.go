package postgres

import (
	"social_server/internal/models/postgres"
	"time"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *sessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// Session management
func (r *sessionRepository) CreateSession(session *postgres.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) GetSessionByTokenID(tokenID string) (*postgres.Session, error) {
	var session postgres.Session
	err := r.db.
		Preload("User").
		Where("token_id = ? AND is_active = ? AND expires_at > ?", tokenID, true, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetSessionBySessionID(sessionID string) (*postgres.Session, error) {
	var session postgres.Session
	err := r.db.
		Preload("User").
		Where("session_id = ? AND is_active = ? AND expires_at > ?", sessionID, true, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetUserSessions(userID uint) ([]*postgres.Session, error) {
	var sessions []*postgres.Session
	err := r.db.
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Order("last_activity DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) UpdateSession(session *postgres.Session) error {
	return r.db.Save(session).Error
}

func (r *sessionRepository) DeactivateSession(tokenID string) error {
	return r.db.
		Model(&postgres.Session{}).
		Where("token_id = ?", tokenID).
		Update("is_active", false).Error
}

func (r *sessionRepository) DeactivateUserSessions(userID uint) error {
	return r.db.
		Model(&postgres.Session{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false).Error
}

func (r *sessionRepository) DeleteExpiredSessions() error {
	return r.db.
		Where("expires_at < ? OR is_active = ?", time.Now(), false).
		Delete(&postgres.Session{}).Error
}

// Token blacklist management
func (r *sessionRepository) BlacklistToken(tokenID string, userID uint, reason string, expiresAt time.Time) error {
	blacklist := &postgres.TokenBlacklist{
		TokenID:   tokenID,
		UserID:    userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return r.db.Create(blacklist).Error
}

func (r *sessionRepository) IsTokenBlacklisted(tokenID string) (bool, error) {
	var count int64
	err := r.db.
		Model(&postgres.TokenBlacklist{}).
		Where("token_id = ? AND expires_at > ?", tokenID, time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *sessionRepository) CleanupExpiredBlacklist() error {
	return r.db.
		Where("expires_at < ?", time.Now()).
		Delete(&postgres.TokenBlacklist{}).Error
}

// Security logging
func (r *sessionRepository) LogSecurityEvent(log *postgres.SecurityLog) error {
	return r.db.Create(log).Error
}

func (r *sessionRepository) GetUserSecurityLogs(userID uint, limit int) ([]*postgres.SecurityLog, error) {
	var logs []*postgres.SecurityLog
	query := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

func (r *sessionRepository) GetHighRiskEvents(since time.Time) ([]*postgres.SecurityLog, error) {
	var logs []*postgres.SecurityLog
	err := r.db.
		Where("risk IN (?, ?) AND created_at > ?", postgres.RiskHigh, postgres.RiskCritical, since).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// Security analysis
func (r *sessionRepository) GetSuspiciousActivities(userID uint, since time.Time) ([]*postgres.SecurityLog, error) {
	var logs []*postgres.SecurityLog
	err := r.db.
		Where("user_id = ? AND event_type = ? AND created_at > ?", userID, postgres.EventSuspiciousActivity, since).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

func (r *sessionRepository) CountFailedLogins(userID uint, since time.Time) (int64, error) {
	var count int64
	err := r.db.
		Model(&postgres.SecurityLog{}).
		Where("user_id = ? AND event_type = ? AND risk IN (?, ?) AND created_at > ?",
			userID, postgres.EventLogin, postgres.RiskHigh, postgres.RiskCritical, since).
		Count(&count).Error
	return count, err
}

func (r *sessionRepository) GetActiveSessionsCount(userID uint) (int64, error) {
	var count int64
	err := r.db.
		Model(&postgres.Session{}).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Count(&count).Error
	return count, err
}
