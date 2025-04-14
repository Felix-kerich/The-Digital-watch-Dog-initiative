package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole defines the role of a user in the system
type UserRole string

const (
	// RoleAdmin represents system administrators
	RoleAdmin UserRole = "ADMIN"
	// RoleAuditor represents financial auditors
	RoleAuditor UserRole = "AUDITOR"
	// RoleFinanceOfficer represents financial officers who can create and manage transactions
	RoleFinanceOfficer UserRole = "FINANCE_OFFICER"
	// RoleManager represents entity managers who can approve transactions
	RoleManager UserRole = "MANAGER"
	// RolePublic represents public users with read-only access to public data
	RolePublic UserRole = "PUBLIC"
)

// User represents a system user
type User struct {
	ID               string         `json:"id" gorm:"primaryKey;type:char(36)"`
	Name             string         `json:"name"`
	Email            string         `json:"email" gorm:"unique"`
	PasswordHash     string         `json:"-"` // Password hash is not exposed in JSON
	Role             UserRole       `json:"role"`
	EntityID         string         `json:"entityId" gorm:"type:char(36);index"` // User may be associated with an entity
	IsActive         bool           `json:"isActive" gorm:"default:true"`
	LastLogin        *time.Time     `json:"lastLogin" gorm:"null"`
	EmailVerified    bool           `json:"emailVerified" gorm:"default:false"`
	TwoFactorEnabled bool           `json:"twoFactorEnabled" gorm:"default:false"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

// UserSession represents a logged-in session
type UserSession struct {
	ID           string     `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID       string     `json:"userId" gorm:"type:char(36);index"`
	Token        string     `json:"token"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	IP           string     `json:"ip"`
	UserAgent    string     `json:"userAgent"`
	CreatedAt    time.Time  `json:"createdAt"`
	RevokedAt    *time.Time `json:"revokedAt" gorm:"null"`
}

// AuditLog represents a log of user actions in the system
type AuditLog struct {
	ID         string    `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID     string    `json:"userId" gorm:"type:char(36);index"`
	Action     string    `json:"action"`
	EntityType string    `json:"entityType"` // What type of entity was acted upon (Transaction, Fund, etc.)
	EntityID   string    `json:"entityId"`   // ID of the entity that was acted upon
	Detail     string    `json:"detail"`     // Additional details about the action
	IP         string    `json:"ip"`
	Timestamp  time.Time `json:"timestamp"`
}

// BeforeCreate will set a UUID rather than numeric ID for User
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	return
}

// BeforeCreate will set a UUID rather than numeric ID for UserSession
func (s *UserSession) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.NewString()
	return
}

// BeforeCreate will set a UUID rather than numeric ID for AuditLog
func (a *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.NewString()
	return
}
