package repository

import (
	"errors"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormSessionRepository implements SessionRepository with GORM
type GormSessionRepository struct {
	DB *gorm.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository() SessionRepository {
	return &GormSessionRepository{
		DB: utils.DB,
	}
}

// Create adds a new session to the database
func (r *GormSessionRepository) Create(session *models.UserSession) error {
	return r.DB.Create(session).Error
}

// FindByID retrieves a session by ID
func (r *GormSessionRepository) FindByID(id string) (*models.UserSession, error) {
	var session models.UserSession
	if err := r.DB.First(&session, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// FindByToken retrieves a session by access token
func (r *GormSessionRepository) FindByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	if err := r.DB.First(&session, "token = ? AND revoked_at IS NULL AND expires_at > ?", token, time.Now()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found or expired")
		}
		return nil, err
	}
	return &session, nil
}

// FindByRefreshToken retrieves a session by refresh token
func (r *GormSessionRepository) FindByRefreshToken(refreshToken string) (*models.UserSession, error) {
	var session models.UserSession
	if err := r.DB.First(&session, "refresh_token = ? AND revoked_at IS NULL", refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found or revoked")
		}
		return nil, err
	}
	return &session, nil
}

// RevokeByUserID revokes all sessions for a user
func (r *GormSessionRepository) RevokeByUserID(userID string) error {
	now := time.Now()
	return r.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// RevokeByToken revokes a session by its token
func (r *GormSessionRepository) RevokeByToken(token string) error {
	now := time.Now()
	return r.DB.Model(&models.UserSession{}).
		Where("token = ? AND revoked_at IS NULL", token).
		Update("revoked_at", now).Error
}

// RevokeExpiredSessions revokes all expired sessions
func (r *GormSessionRepository) RevokeExpiredSessions() error {
	now := time.Now()
	return r.DB.Model(&models.UserSession{}).
		Where("expires_at < ? AND revoked_at IS NULL", now).
		Update("revoked_at", now).Error
}
