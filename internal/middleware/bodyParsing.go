package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/models"
)

const NotifBodyContextKey string = "notificationBody"

func BodyParsingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Body == nil {
			return customerrors.ErrBadRequest
		}

		var notification models.Notification
		if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
			return customerrors.ErrBadRequest
		}

		if notification.Receiver == "" || notification.Message == "" {
			return customerrors.ErrBadRequest
		}

		sender, err := GetUsernameFromContext(r.Context())
		if err != nil {
			return err
		}
		notification.Sender = sender

		ctx := context.WithValue(r.Context(), NotifBodyContextKey, &notification)
		*r = *r.WithContext(ctx)

		return next(w, r)
	}
}

func GetNotificationFromContext(ctx context.Context) (*models.Notification, error) {
	notificationVal := ctx.Value(NotifBodyContextKey)
	notification, ok := notificationVal.(*models.Notification)
	if !ok || notification == nil {
		return nil, customerrors.ErrBadRequest
	}
	return notification, nil
}
