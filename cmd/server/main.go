package main

import (
	"context"

	"github.com/taekwondodev/push-notification-service/internal/api"
	"github.com/taekwondodev/push-notification-service/internal/config"
	"github.com/taekwondodev/push-notification-service/internal/controller"
	"github.com/taekwondodev/push-notification-service/internal/repository"
	"github.com/taekwondodev/push-notification-service/internal/service"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

func main() {
	cfg := config.Load()

	repo := repository.NewMongoNotificationRepository(cfg.Mongo.URI, cfg.Mongo.Database)
	defer repo.Close()

	notifService := service.NewNotificationService(repo)
	hub := websocket.NewHub()
	kafkaService := service.NewKafkaService(&cfg.Kafka, hub, notifService)

	notifController := controller.NewNotificationController(notifService, kafkaService)
	wsController := controller.NewWebSocketController(hub)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := kafkaService.StartConsumer(ctx); err != nil {
			cancel()
		}
	}()

	router := api.SetupRoutes(notifController, wsController)
	server := api.NewServer(cfg.Server.Port, router)
	server.StartWithGracefulShutdown()
}
