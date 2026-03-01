package database

import (
	"context"
	"fmt"
	"log"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectRedis(cfg *config.Config) *redis.Client {
	redisHost := cfg.RedisHost
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := cfg.RedisPort
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := cfg.RedisPassword // usually empty for local dev

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0, // use default DB
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("[Warning] Failed to connect to Redis: %v. JWT Revocation features may not function.", err)
	} else {
		log.Println("Connected to Redis successfully")
	}

	RedisClient = rdb
	return rdb
}
