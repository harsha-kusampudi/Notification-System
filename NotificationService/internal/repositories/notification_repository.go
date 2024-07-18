package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"

	"project/notification/internal/models"
)

type NotificationRepositoryInterface interface {
	AddNotification(ctx context.Context, notification models.Notification) error
	GetNotificationsInRange(ctx context.Context, startTime, endTime string) (*[]models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, notification models.Notification) error
}

type NotificationRepository struct {
	dbClient *dynamodb.Client
}

func (repo *NotificationRepository) AddNotification(ctx context.Context,notification models.Notification) error {
	_, err := repo.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("Notifications"), // Replace with your table name
		Item: map[string]types.AttributeValue{
			"Msg": &types.AttributeValueMemberS{Value: notification.Message},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s#%s", notification.NotificationTime, notification.Id)},
			"PK": &types.AttributeValueMemberS{Value: "notification"},
			"Delivered": &types.AttributeValueMemberBOOL{Value: notification.Delivered},
		},
	})

	if err != nil {
		fmt.Printf("Failed to add item to DynamoDB: %s\n", err)
		return err
	}

	fmt.Println("Added item to DynamoDB table.")
	return nil
}

func (repo *NotificationRepository) GetNotificationsInRange(ctx context.Context, startTime, endTime string) (*[]models.Notification, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String("Notifications"),
        KeyConditionExpression: aws.String("PK = :type AND #ts BETWEEN :start AND :end"),
        ExpressionAttributeNames: map[string]string{
            "#ts": "SK",
        },
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":type":  &types.AttributeValueMemberS{Value: "notification"},
            ":start": &types.AttributeValueMemberS{Value: startTime},
            ":end":   &types.AttributeValueMemberS{Value: endTime + "#ffffffff-ffff-ffff-ffff-ffffffffffff"},
        },
    }

    result, err := repo.dbClient.Query(context.TODO(), input)
    if err != nil {
		fmt.Println(err)
        return nil, fmt.Errorf("failed to query notifications: %v", err)
    }

    var notifications []models.Notification
	for _, item := range result.Items {
		split := strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")
		notification := models.Notification{
			Message: item["Msg"].(*types.AttributeValueMemberS).Value,
			NotificationTime: split[0],
			Id: split[1],
			Delivered: item["Delivered"].(*types.AttributeValueMemberBOOL).Value,
		}
		notifications = append(notifications, notification)
	}
    return &notifications, nil
}

func (repo *NotificationRepository) UpdateNotificationStatus(ctx context.Context, notification models.Notification) error {
	_, err := repo.dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("Notifications"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "notification"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s#%s", notification.NotificationTime, notification.Id)},
		},
		UpdateExpression: aws.String("SET Delivered = :d"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":d": &types.AttributeValueMemberBOOL{Value: notification.Delivered},
		},
	})

	if err != nil {
		fmt.Printf("Failed to update item in DynamoDB: %s\n", err)
		return err
	}

	fmt.Println("Updated item in DynamoDB table.")
	return nil
}

func NewNotificationRepository(dbClient *dynamodb.Client) NotificationRepositoryInterface {
	return &NotificationRepository{
		dbClient: dbClient,
	}
}