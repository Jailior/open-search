package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis interface
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// Makes a default RedisClient, gets Redis addres from environment
func MakeRedisClient() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	return &RedisClient{
		Client: rdb,
		Ctx:    context.TODO(),
	}
}

// Clears stream with streamName
func (r *RedisClient) ResetStream(streamName string) {
	r.Client.Del(r.Ctx, streamName)
	log.Println("Reset Redis stream")
}

// Clears queue and set with respective names
func (r *RedisClient) ResetQueueAndSet(queueName, setName string) {
	r.Client.Del(r.Ctx, queueName)
	r.Client.Del(r.Ctx, setName)
	log.Println("Reset Redis queue and set")
}

// Ensures a Redis Consumer Group is available in the stream
func (r *RedisClient) EnsureConsumerGroup(streamName, group string) {
	err := r.Client.XGroupCreateMkStream(r.Ctx, streamName, group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatalf("Failed to create consumer group: %v\n", err)
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
func (r *RedisClient) ReadStream(streamName string, group string, consumerName string) ([]redis.XMessage, error) {
	readIDs := []string{"0", ">"}

	// "0" reads any unacknowledged messages
	// "0" reads any new messages

	for _, id := range readIDs {

		entries, err := r.Client.XReadGroup(r.Ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumerName,
			Streams:  []string{streamName, id},
			Count:    10,
			Block:    5 * time.Second, // blocks for 5 seconds
		}).Result()

		if err == redis.Nil {
			// timeout likely
			continue
		}

		if err != nil {
			log.Println("Redis XRead Error: ", err)
			return nil, err
		}

		if len(entries) == 0 || len(entries[0].Messages) == 0 {
			// return nil, fmt.Errorf("Error reading from stream, internal")
			continue
		}

		return entries[0].Messages, nil
	}

	// no messages found
	return nil, nil
}

// Enqueues url to list if url is not in set
func (r *RedisClient) EnqueueList(url, listName, setName string) error {
	err := r.Client.RPush(r.Ctx, listName, url).Err()
	if err != nil {
		return fmt.Errorf("Failed to enqueue page to list: %v\n", err)
	}
	return nil
}

// Enqueues url to front of list
func (r *RedisClient) EnqueueFront(url, listName string) error {
	err := r.Client.RPush(r.Ctx, listName, url).Err()
	if err != nil {
		return fmt.Errorf("Failed to enqueue page to list: %v\n", err)
	}
	return nil
}

// Dequeues url from list
func (r *RedisClient) DequeueList(listName string) (string, error) {
	ctx, cancel := context.WithTimeout(r.Ctx, 5*time.Second)
	defer cancel()

	url, err := r.Client.BLPop(ctx, 5*time.Second, listName).Result()
	if err == redis.Nil || len(url) < 2 {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("Redis BLPop error: %w", err)
	}
	return url[1], nil
}

// Adds url to set
func (r *RedisClient) SetAdd(url, setName string) {
	r.Client.SAdd(r.Ctx, setName, url)
}

// Returns true if url is present in set
func (r *RedisClient) SetHas(url, setName string) bool {
	exists, _ := r.Client.SIsMember(r.Ctx, setName, url).Result()
	return exists
}
