package controller

import (
	"encoding/json"
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/middleware"
	"github.com/taekwondodev/push-notification-service/internal/service"
)

type NotificationController struct {
	notifSvc *service.NotificationService
	kafkaSvc *service.KafkaService
}

func NewNotificationController(notifSvc *service.NotificationService, kafkaSvc *service.KafkaService) *NotificationController {
	return &NotificationController{
		notifSvc: notifSvc,
		kafkaSvc: kafkaSvc,
	}
}

func (c *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) error {
	username, err := middleware.GetUsernameFromContext(r.Context())
	if err != nil {
		return err
	}
	unreadOnly, err := middleware.GetUnreadFromContext(r.Context())
	if err != nil {
		return err
	}

	notifications, err := c.notifSvc.GetNotificationsByReceiver(r.Context(), username, unreadOnly)
	if err != nil {
		return err
	}

	return c.writeResponse(w, http.StatusOK, notifications)
}

func (c *NotificationController) CreateNotification(w http.ResponseWriter, r *http.Request) error {
	notification, err := middleware.GetNotificationFromContext(r.Context())
	if err != nil {
		return err
	}

	if err := c.kafkaSvc.PublishNotification(r.Context(), notification); err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (c *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return customerrors.ErrBadRequest
	}

	if err := c.notifSvc.MarkAsRead(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *NotificationController) writeResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
