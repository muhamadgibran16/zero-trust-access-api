package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gibran/go-gin-boilerplate/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// Connect to the default "postgres" database first
	encodedPass := url.QueryEscape(cfg.DBPassword)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		cfg.DBUser, encodedPass, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)

	log.Printf("Connecting to postgres to create database: %s", cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to postgres server: %v", err)
	}

	// Execute CREATE DATABASE
	err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\";", cfg.DBName)).Error
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	log.Println("Database successfully created!")
}
