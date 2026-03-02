package main

import (
	"log"
	"time"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	apps := []model.AppRoute{
		{
			ID:          uuid.New(),
			Name:        "HR Information System",
			Description: "Internal portal for employee data and payroll. (Proxy dummy target)",
			PathPrefix:  "/hr-app",
			TargetURL:   "http://localhost:9090",
			Icon:        "👥",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Financial Dashboard",
			Description: "Secure finance and accounting dashboard.",
			PathPrefix:  "/finance",
			TargetURL:   "http://localhost:9091",
			Icon:        "💰",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Engineering Wiki",
			Description: "Internal engineering documentation and wiki.",
			PathPrefix:  "/eng-wiki",
			TargetURL:   "http://localhost:9092",
			Icon:        "📘",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, app := range apps {
		if err := db.FirstOrCreate(&app, model.AppRoute{PathPrefix: app.PathPrefix}).Error; err != nil {
			log.Printf("Failed to insert %s: %v", app.Name, err)
		} else {
			log.Printf("Inserted/Verified app: %s", app.Name)
		}
	}

	log.Println("Database seeded with AppRoutes")
}
