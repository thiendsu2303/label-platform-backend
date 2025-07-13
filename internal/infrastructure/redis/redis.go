package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	QueueGPT    = "label-platform-queue-gpt"
	QueueClaude = "label-platform-queue-claude"
	QueueGemini = "label-platform-queue-gemini"
	QueueResult = "label-platform-queue-result"
)

// RedisClient wraps the go-redis client
var RedisClient *redis.Client

// NewRedisConnection initializes the Redis connection and creates the required queues
func NewRedisConnection(ctx context.Context) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_HOST", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	// Test connection
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Ensure the 4 queues exist (create empty lists if not exist)
	queues := []string{QueueGPT, QueueClaude, QueueGemini, QueueResult}
	for _, q := range queues {
		// Use RPush to create the list if it doesn't exist, then LPop to remove the dummy value
		if err := RedisClient.RPush(ctx, q, "__init__").Err(); err != nil {
			return fmt.Errorf("failed to create queue %s: %w", q, err)
		}
		RedisClient.LPop(ctx, q)
	}

	fmt.Println("[Redis] Connected and initialized queues:", queues)
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
