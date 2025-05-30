package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
)

const UnreadContextKey string = "unread"

func QueryParsingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var unread *bool
		if unreadStr := r.URL.Query().Get("unread"); unreadStr != "" {
			unreadOnly, err := strconv.ParseBool(unreadStr)
			if err != nil {
				return customerrors.ErrBadRequest
			}
			unread = &unreadOnly
		}

		ctx := context.WithValue(r.Context(), UnreadContextKey, unread)
		*r = *r.WithContext(ctx)

		return next(w, r)
	}
}

func GetUnreadFromContext(ctx context.Context) (bool, error) {
	unreadVal := ctx.Value(UnreadContextKey)
	unread, ok := unreadVal.(*bool)
	if !ok || unread == nil {
		return false, customerrors.ErrBadRequest
	}
	return *unread, nil
}
