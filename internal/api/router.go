package api

import (
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/controller"
	"github.com/taekwondodev/push-notification-service/internal/middleware"
)

var router *http.ServeMux

func SetupRoutes(notifC *controller.NotificationController, wsC *controller.WebSocketController) *http.ServeMux {
	router = http.NewServeMux()

	router.Handle("OPTIONS /", middleware.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	setupNotificationRoutes(notifC)
	setupWSRoutes(wsC)

	return router
}

func setupNotificationRoutes(notifC *controller.NotificationController) {
	router.Handle("POST /notifications", applyPostMiddleware(notifC.CreateNotification))
	router.Handle("GET /notifications", applyGetMiddleware(notifC.GetNotifications))
	router.Handle("PATCH /notifications/{id}", applyMiddleware(notifC.MarkAsRead))
}

func setupWSRoutes(wsC *controller.WebSocketController) {
	router.Handle("GET /ws", applyWSMiddleware(wsC.HandleConnection))
}

func applyMiddleware(h middleware.HandlerFunc) http.HandlerFunc {
	return middleware.CorsMiddleware(
		middleware.ErrorHandler(
			middleware.LoggingMiddleware(
				middleware.AuthMiddleware(h),
			),
		),
	)
}

func applyPostMiddleware(h middleware.HandlerFunc) http.HandlerFunc {
	return middleware.CorsMiddleware(
		middleware.ErrorHandler(
			middleware.LoggingMiddleware(
				middleware.AuthMiddleware(
					middleware.BodyParsingMiddleware(h),
				),
			),
		),
	)
}

func applyGetMiddleware(h middleware.HandlerFunc) http.HandlerFunc {
	return middleware.CorsMiddleware(
		middleware.ErrorHandler(
			middleware.LoggingMiddleware(
				middleware.AuthMiddleware(
					middleware.QueryParsingMiddleware(h),
				),
			),
		),
	)
}

func applyWSMiddleware(h middleware.HandlerFunc) http.HandlerFunc {
	return middleware.CorsMiddleware(
		middleware.ErrorHandler(
			middleware.LoggingMiddleware(h),
		),
	)
}
