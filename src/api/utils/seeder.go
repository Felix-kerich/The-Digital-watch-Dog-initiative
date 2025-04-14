package utils

import (
	"log"
	"os"

	"github.com/the-digital-watchdog-initiative/models"
	"gorm.io/gorm"
)

// SeedAdmin creates an admin user if it doesn't exist
func SeedAdmin(db *gorm.DB) error {
	// Check if admin exists
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)
	if count > 0 {
		log.Println("Admin user already exists")
		return nil
	}

	// Get admin credentials from environment variables
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminName := os.Getenv("ADMIN_NAME")

	if adminEmail == "" || adminPassword == "" || adminName == "" {
		return nil // Skip if env variables are not set
	}

	// Hash password
	hashedPassword, err := GeneratePasswordHash(adminPassword)
	if err != nil {
		return err
	}

	// Create admin user
	admin := &models.User{
		Name:         adminName,
		Email:        adminEmail,
		PasswordHash: hashedPassword,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	result := db.Create(admin)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Admin user created successfully: %s\n", adminEmail)
	return nil
}
