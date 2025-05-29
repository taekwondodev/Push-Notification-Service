package api

import (
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/controller"
	"github.com/taekwondodev/push-notification-service/internal/middleware"
)

var router *http.ServeMux

func SetupRoutes(notifC *controller.NotificationController, wsC *controller.WebSocketController) *http.ServeMux {
	router = http.NewServeMux()

	router.Handle("/", middleware.CorsMiddleware(router))
	setupNotificationRoutes(notifC)
	setupWSRoutes(wsC)

	return router
}

func applyMiddleware(h middleware.HandlerFunc) http.HandlerFunc {
	handlerWithLogging := middleware.LoggingMiddleware(h)
	errorHandler := middleware.ErrorHandler(handlerWithLogging)
	return middleware.CorsMiddleware(errorHandler)
}

func setupNotificationRoutes(notifC *controller.NotificationController) {
	router.Handle("POST /notifications", applyMiddleware(notifC.CreateNotification))
	router.Handle("GET /notifications", applyMiddleware(notifC.GetNotifications))
	router.Handle("POST /notifications/{id}/read", applyMiddleware(notifC.MarkAsRead))
}

func setupWSRoutes(wsC *controller.WebSocketController) {
	router.Handle("GET /ws", applyMiddleware(wsC.HandleConnection))
}