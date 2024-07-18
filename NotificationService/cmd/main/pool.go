package main

import (
	"context"
	"fmt"
	"project/notification/internal/controllers"
	"project/notification/internal/models"
	"sync"
)

type WorkerPool struct {
	wg *sync.WaitGroup
	channel chan models.Notification
	workers int
	controller *controllers.NotificationController
	ctx context.Context
}

func NewWorkerPool(ctx context.Context,workers int, c *controllers.NotificationController) *WorkerPool {
	return &WorkerPool{
		wg: &sync.WaitGroup{},
		channel: make(chan models.Notification),
		workers: workers,
		controller: c,
		ctx: ctx,
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for notification := range wp.channel {
		fmt.Printf("Processing notification: %s\n", notification.Message)
		// time.Sleep(3 * time.Second) // Time to process, eg: send an email
		notification.Delivered = true
		wp.controller.UpdateNotificationStatus(wp.ctx, notification)
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}


