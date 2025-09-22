package configs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	rdbHost := os.Getenv("RDBHOST")
	rdbPort := os.Getenv("RDBPORT")
	rdbPass := os.Getenv("RDBPASS")
	rdbUser := os.Getenv("RDBUSER")

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdbHost, rdbPort),
		Password: rdbPass,
		Username: rdbUser,
	})

	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v\n", err)
		return nil, err
	}
	log.Printf("Redis Connected")
	return client, nil
}
