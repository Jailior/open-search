package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // e.g., "localhost:6379"
		Password: "",
		DB:       0,
	})

	keys := []string{"url_queue", "visited_set", "pages_to_index"}
	for _, key := range keys {
		err := rdb.Del(ctx, key).Err()
		if err != nil {
			fmt.Printf("Failed to delete key %s: %v\n", key, err)
		} else {
			fmt.Printf("Deleted key: %s\n", key)
		}
	}
}
