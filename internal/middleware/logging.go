package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
)

func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		err := next(w, r)

		duration := time.Since(start)
		status := http.StatusOK
		if err != nil {
			status = customerrors.GetStatus(err)
		}

		log.Printf("Completed %s %s | Status: %d | Duration: %v",
			r.Method, r.URL.Path, status, duration)

		return err
	}
}
