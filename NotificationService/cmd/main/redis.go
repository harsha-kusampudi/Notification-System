package main

import (
	"context"
	"fmt"
	"log"
	"project/notification/internal/models"
	"project/notification/utils"
	"github.com/go-redis/redis/v8"
)


/***
    * Used a set to store the notification with the score as the time and member as the notification id
    * Used a hash to store the notification with the notification id as the key and the message and time as the value

    * The AddNotification function adds the notification to the set and the hash
*/


type redisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context) *redisClient {
	// Create a new redis client

	redisclient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Redis server address
        Password: "",               // No password set
        DB:       0,                // Use default DB
    })

    // Ping the Redis server to check if the connection is alive
    pong, err := redisclient.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Could not connect to Redis: %v", err)
    }
    fmt.Println("Connected to Redis:", pong)
	return &redisClient{
		client: redisclient,
	}
}

func (c *redisClient) AddNotification(ctx context.Context, notification models.Notification) error {
    // parse time
	if notification.Delivered {
        fmt.Printf("Notification already delivered: %v\n", notification)
        return nil
    }
	time, err := utils.TimeStringToUnix(notification.NotificationTime)
	if err != nil {
		return err
	}
    err = c.client.ZAdd(ctx, "notifications", &redis.Z{
        Score:  float64(time),
        Member: notification.Id,
    }).Err()
    if err != nil {
        return err
    }
	fmt.Printf("Added notification to Redis: %v\n", notification)
    
    return c.client.HSet(ctx, "notification:"+notification.Id, map[string]interface{}{
        "message": notification.Message,
        "time": notification.NotificationTime,
    }).Err()

}
