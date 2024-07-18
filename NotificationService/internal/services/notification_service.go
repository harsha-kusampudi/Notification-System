package services

import (
	"context"

	"project/notification/internal/models"
	"project/notification/internal/repositories"
)

type NotificationServiceInterface interface {
	AddNotification(ctx context.Context, notification models.Notification) error
	GetNotificationsInRange(ctx context.Context, startTime, endTime string) (*[]models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, notification models.Notification) error
}

type NotificationService struct {
	notificationRepo repositories.NotificationRepositoryInterface
}

func (notificationService *NotificationService) AddNotification(ctx context.Context, notification models.Notification) error {
	err := notificationService.notificationRepo.AddNotification(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func (notificationService *NotificationService) GetNotificationsInRange(ctx context.Context, startTime, endTime string) (*[]models.Notification, error) {
	notifications, err := notificationService.notificationRepo.GetNotificationsInRange(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (notificationService *NotificationService) UpdateNotificationStatus(ctx context.Context, notification models.Notification) error {
	err := notificationService.notificationRepo.UpdateNotificationStatus(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func NewNotificationService(notificationRepo repositories.NotificationRepositoryInterface) NotificationServiceInterface {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}
