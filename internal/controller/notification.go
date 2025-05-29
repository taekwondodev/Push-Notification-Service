package controller

import (
	"encoding/json"
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/service"
)

type NotificationController struct {
    notifSvc  *service.NotificationService
    kafkaSvc  *service.KafkaService
}

func NewNotificationController(notifSvc *service.NotificationService, kafkaSvc *service.KafkaService) *NotificationController {
    return &NotificationController{
        notifSvc: notifSvc,
        kafkaSvc: kafkaSvc,
    }
}

func (h *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) error {
    user := r.URL.Query().Get("user")
    if user == "" {
        return customerrors.ErrBadRequest
    }

    notifications, err := h.notifSvc.GetNotificationsByReceiver(r.Context(), user)
    if err != nil {
        return err
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    return json.NewEncoder(w).Encode(notifications)
}

func (h *NotificationController) CreateNotification(w http.ResponseWriter, r *http.Request) error {
    var notification models.Notification
    if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
        return customerrors.ErrBadRequest
    }

    if err := h.kafkaSvc.PublishNotification(r.Context(), &notification); err != nil {
        return err
    }

    w.WriteHeader(http.StatusAccepted)
    return nil
}

func (h *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) error {
    id := r.PathValue("id")
    if id == "" {
        return customerrors.ErrBadRequest
    }

    if err := h.notifSvc.MarkAsRead(r.Context(), id); err != nil {
        return err
    }

    w.WriteHeader(http.StatusOK)
    return nil
}