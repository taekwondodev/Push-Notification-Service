package middleware

import (
	"context"
	"net/http"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
)

const UserContextKey string = "username"

func AuthMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, err := extractUserFromHeaders(r)
		if err != nil {
			return err
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		*r = *r.WithContext(ctx)

		return next(w, r)
	}
}

func extractUserFromHeaders(r *http.Request) (string, error) {
	username := r.Header.Get("X-User-Username")

	if username == "" {
		return "", customerrors.ErrBadRequest
	}

	return username, nil
}

func GetUsernameFromContext(ctx context.Context) (string, error) {
	usernameVal := ctx.Value(UserContextKey)
	username, ok := usernameVal.(string)
	if !ok {
		return "", customerrors.ErrBadRequest
	}
	return username, nil
}
