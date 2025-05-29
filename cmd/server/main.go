package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/config"
	"github.com/taekwondodev/push-notification-service/internal/controller"
	"github.com/taekwondodev/push-notification-service/internal/repository"
	"github.com/taekwondodev/push-notification-service/internal/service"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

func main() {
    cfg := config.Load()

    repo, err := repository.NewMongoNotificationRepository(cfg.Mongo.URI, cfg.Mongo.Database)
    if err != nil {
        log.Fatal("failed to connect to database", "error", err)
    }
    defer repo.Close()
    
    notifService := service.NewNotificationService(repo)
    hub := websocket.NewHub()
    kafkaService := service.NewKafkaService(&cfg.Kafka, hub, notifService)
    
    notifController := controller.NewNotificationController(notifService, kafkaService)
    wsController := controller.NewWebSocketController(hub)
    
    ctx, cancel := context.WithCancel(context.Background())
    go kafkaService.StartConsumer(ctx)
    
    router := http.NewServeMux()
    router.HandleFunc("GET /ws", wsController.HandleConnection)
    router.HandleFunc("GET /notifications", notifController.GetNotifications)
    router.HandleFunc("POST /notifications", notifController.CreateNotification)
    router.HandleFunc("POST /notifications/{id}/read", notifController.MarkAsRead)
    
    server := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
    }
    
    go func() {
        log.Println("server starting", "port", cfg.Server.Port)
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal("server failed", "error", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("shutting down server...")
    
    cancel()
    ctx, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancelShutdown()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("server forced to shutdown", "error", err)
    }
    
    log.Println("server exited")
}