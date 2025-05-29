package controller

import (
	"encoding/json"
	"net/http"

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

func (h *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) {
    user := r.URL.Query().Get("user")
    if user == "" {
        http.Error(w, "user parameter required", http.StatusBadRequest)
        return
    }

    notifications, err := h.notifSvc.GetNotificationsByReceiver(r.Context(), user)
    if err != nil {
        http.Error(w, "failed to fetch notifications", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(notifications); err != nil {
		// return err
    }
}

func (h *NotificationController) CreateNotification(w http.ResponseWriter, r *http.Request) {
    var notification models.Notification
    if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    if err := h.kafkaSvc.PublishNotification(r.Context(), &notification); err != nil {
        http.Error(w, "failed to publish notification", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusAccepted)
}

func (h *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        http.Error(w, "id parameter required", http.StatusBadRequest)
        return
    }

    if err := h.notifSvc.MarkAsRead(r.Context(), id); err != nil {
        http.Error(w, "failed to mark notification as read", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}