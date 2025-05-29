package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

func MakeRedisClient() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	return &RedisClient{
		Client: rdb,
		Ctx:    context.Background(),
	}
}

// Pushes a key value pair to a stream
func (r *RedisClient) PushToStream(stream string, key string, value string) error {
	err := r.Client.XAdd(r.Ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{key: value},
	}).Err()
	return err
}

// Reads a key value pair from a stream
func (r *RedisClient) ReadAckStream(streamName string, group string, consumerName string) ([]redis.XStream, error) {
	entries, err := r.Client.XReadGroup(r.Ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumerName,
		Streams:  []string{streamName, ">"},
		Count:    10,
		Block:    0, // blocks forever
	}).Result()

	if err != nil {
		fmt.Println("Redis XRead Error: ", err)
		return nil, err
	}

	return entries, nil
}
