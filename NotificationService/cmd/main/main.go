package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"project/notification/internal/controllers"
	"project/notification/internal/models"
	"project/notification/internal/repositories"
	"project/notification/internal/services"
	"project/notification/utils"
	"time"
	"flag"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/cors"
)

func handleInput() (string, string, int) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Printf("Press 1 to Enter message and notification time, 2 to exit: ")
    var action int
    _, err := fmt.Scanf("%d\n", &action)
    if err != nil {
        fmt.Println("Error reading action:", err)
        return "", "", 1
    }

    switch action {
    case 1:
        fmt.Printf("Enter message: ")
        message, _ := reader.ReadString('\n')
        message = message[:len(message)-1]

        fmt.Printf("Enter notification time (e.g., 2024-07-14 7:29:00): ")
        notificationTime, _ := reader.ReadString('\n')
        notificationTime = notificationTime[:len(notificationTime)-1]

        fmt.Printf("Message: %s, Notification Time: %s\n", message, notificationTime)

        return message, notificationTime, 0
    case 2:
        fmt.Println("Exiting...")
        return "", "", 1

    default:
        fmt.Println("Invalid action.")
        return "", "", 1
    }
}

// Add notifications from dynamodb to redis everyday
func D2R_everyday(ctx context.Context, c *controllers.NotificationController, rc *redisClient){
	
	for {
		// Get current time in IST
		now := time.Now().In(utils.ISTLocation)

		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.ISTLocation)
		endOfDay := startOfDay.Add(24*time.Hour - time.Second)

		startOfDayStr := startOfDay.Format("2006-01-02 15:04:05")
		endOfDayStr := endOfDay.Format("2006-01-02 15:04:05")
		fmt.Printf("Start of day: %s\n", endOfDayStr)

		notifications, err := c.GetNotificationsInRange(ctx, startOfDayStr, endOfDayStr)
		if err != nil {
			fmt.Println("Failed to get notifications from DynamoDB",err)
		} else{
			for _, notification := range *notifications {
				if notification.Delivered {
					continue
				}
				err := rc.AddNotification(ctx, notification)
				if err != nil {
					fmt.Println("Failed to add notification to Redis",err)
				}

			}
		}
    	// Calculate next day's date by adding 1 day to the current date and truncating time to midnight
		nextDay := time.Date(now.Year(), now.Month(), now.Day() + 1, 0, 0, 0, 0, utils.ISTLocation)
		// Calculate duration until next day's midnight
		durationUntilNextDay := nextDay.Sub(time.Now().In(utils.ISTLocation))

		fmt.Printf("Sleeping for %v until next day\n", durationUntilNextDay)
		time.Sleep(durationUntilNextDay)
		
	}
}
// adds notification into dynamodb (and redis if notification time is today)
func AddNotification(ctx context.Context, c *controllers.NotificationController, rc *redisClient, notification models.Notification) {
	c.AddNotification(ctx, notification)

	notificationTimeParsed, _ := time.ParseInLocation("2006-01-02 15:04:05", notification.NotificationTime, utils.ISTLocation)
	fmt.Printf("Notification will be sent at: %s\n", notificationTimeParsed)

	// if notificationTimeParsed is today 
	now := time.Now().In(utils.ISTLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	notificationDate := time.Date(notificationTimeParsed.Year(), notificationTimeParsed.Month(), notificationTimeParsed.Day(), 0, 0, 0, 0, notificationTimeParsed.Location())

	if notificationDate.Equal(today) {
		// store in redis
		err := rc.AddNotification(ctx, notification)
		if err != nil {
			fmt.Println("Failed to add notification to Redis",err)
		}
	}
}

func run_cli(ctx context.Context, c *controllers.NotificationController, rc *redisClient) {
	// Run the program
	for {
		message, notificationTime, status := handleInput()
		
		if status != 0 {
			break
		}

		notification := models.Notification{
			Id: uuid.New().String(),
			Message: message,
			NotificationTime: notificationTime,
		}

		AddNotification(ctx, c, rc, notification)
	}
}

// Continuously checks for any notifications due in redis
func checkForDueNotifications (ctx context.Context, wp *WorkerPool, rc *redisClient) {
	// check for notifications which are due in redis
	// if due, push to worker pool channel

	for {
		notifications, err := rc.client.ZRangeByScore(ctx, "notifications", &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%d", time.Now().In(utils.ISTLocation).Unix()),
		}).Result()
		if err != nil {
			fmt.Println("Failed to get notifications from Redis")
		}

		for _, notificationId := range notifications {
			notification, err := rc.client.HGetAll(ctx, "notification:"+notificationId).Result()
			if err != nil {
				fmt.Println("Failed to get notification from Redis")
			}

			notificationTime, err := time.ParseInLocation("2006-01-02 15:04:05", notification["time"], utils.ISTLocation)
			if err != nil {
				fmt.Println("Failed to parse notification time")
			}
			fmt.Printf("Notification is due: %s\n", notificationTime)
			wp.channel <- models.Notification{
				Id: notificationId,
				Message: notification["message"],
				NotificationTime: notification["time"],
			}
		}
		
		// remove notifications from redis
		for _, notification := range notifications {
			_, err := rc.client.ZRem(ctx, "notifications", notification).Result()
			if err != nil {
				// Handle error
				fmt.Println("Error removing notification:", err)
			}
		}

		for _, notificationId := range notifications {
			rc.client.ZRem(ctx, "notifications", notificationId)
		}


	}
}

func run_web(ctx context.Context, c *controllers.NotificationController, rc *redisClient, notification models.Notification) {
	AddNotification(ctx, c, rc, notification)
}

func init() {
	var err error
	utils.ISTLocation, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading IST location:", err)
		return
	}
}

func main() {

	ctx := context.Background()
	dbClient := NewDynamoDBClient(ctx)
	redisClient := NewRedisClient(ctx)
	repo := repositories.NewNotificationRepository(dbClient)
	service := services.NewNotificationService(repo)
	controller := controllers.NewNotificationController(service)


	WorkerPool := NewWorkerPool(ctx, 5, controller)
	WorkerPool.Start()
	
	go checkForDueNotifications(ctx, WorkerPool, redisClient)
	go D2R_everyday(ctx, controller, redisClient)

	mode := flag.String("mode", "cli", "Application mode: cli or web")

    flag.Parse()
	
	if *mode == "cli" {
		run_cli(ctx, controller, redisClient)
	} else if *mode == "web" {
		mux := http.NewServeMux()

		// Set up your routes
		mux.HandleFunc("/schedule", func (w http.ResponseWriter, r *http.Request) {
			
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
				return
			}
		
			var notification models.Notification
			// Use json.NewDecoder to decode the request body into the struct
			fmt.Println(r.Body)
			err := json.NewDecoder(r.Body).Decode(&notification)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Printf("Received notification: %v\n", notification)
			notification.Id = uuid.New().String()
			notification.Delivered = false
			run_web(ctx, controller, redisClient, notification)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(struct {
				Status string `json:"status"`
			}{
				Status: "Success",
			})

		})


		// Create a CORS middleware
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
			AllowCredentials: true,
		})

		// Wrap your mux with the CORS middleware
		handler := c.Handler(mux)

		// Start the server
		http.ListenAndServe(":8081", handler)
	}else{
		fmt.Println("Invalid mode")
	}

}