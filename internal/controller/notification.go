package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/models"
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
	user, unreadOnly, err := c.parseGetNotificationsRequest(r)
	if err != nil {
		return err
	}

	notifications, err := c.notifSvc.GetNotificationsByReceiver(r.Context(), user, unreadOnly)
	if err != nil {
		return err
	}

	return c.writeResponse(w, http.StatusOK, notifications)
}

func (c *NotificationController) CreateNotification(w http.ResponseWriter, r *http.Request) error {
	notification, err := c.parseCreateNotificationRequest(r)
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

func (c *NotificationController) parseGetNotificationsRequest(r *http.Request) (string, bool, error) {
	user := r.URL.Query().Get("user")
	if user == "" {
		return "", false, customerrors.ErrBadRequest
	}

	var unread *bool
	if unreadStr := r.URL.Query().Get("unread"); unreadStr != "" {
		unreadOnly, err := strconv.ParseBool(unreadStr)
		if err != nil {
			return "", false, customerrors.ErrBadRequest
		}
		unread = &unreadOnly
	}

	return user, *unread, nil
}

func (c *NotificationController) parseCreateNotificationRequest(r *http.Request) (*models.Notification, error) {
	var notification models.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		return nil, customerrors.ErrBadRequest
	}

	if notification.Sender == "" || notification.Receiver == "" || notification.Message == "" {
		return nil, customerrors.ErrBadRequest
	}

	return &notification, nil
}

func (c *NotificationController) writeResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
