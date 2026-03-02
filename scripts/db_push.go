package main

import (
	"log"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db := database.Connect(cfg)

	log.Println("Running AutoMigrate...")
	
	// Auto-migrate models
	err := db.AutoMigrate(
		&model.User{},
		&model.AuditLog{},
		&model.PolicyRule{},
		&model.Device{},
		&model.Notification{},
		&model.PasswordReset{},
		&model.AppRoute{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	log.Println("Database schema successfully pushed and synchronized!")
}
