package utils

import (
    "github.com/go-redis/redis/v8"
)

// InitRedisClient initializes and returns a Redis client
func InitRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Update if Redis is running elsewhere
        Password: "",               // Set if Redis requires a password
        DB:       0,                // Use default DB
    })
}
