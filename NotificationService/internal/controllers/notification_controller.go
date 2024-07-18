package controllers

import (
	"context"
	"project/notification/internal/models"
	"project/notification/internal/services"
)


type NotificationController struct {
	notificationService services.NotificationServiceInterface
}

func (c *NotificationController) AddNotification(ctx context.Context, notification models.Notification) error {
	err := c.notificationService.AddNotification(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func (c *NotificationController) GetNotificationsInRange(ctx context.Context, startTime, endTime string) (*[]models.Notification, error) {
	notifications, err := c.notificationService.GetNotificationsInRange(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (c *NotificationController) UpdateNotificationStatus(ctx context.Context, notification models.Notification) error {
	err := c.notificationService.UpdateNotificationStatus(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func NewNotificationController(notificationService services.NotificationServiceInterface) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
	}
}