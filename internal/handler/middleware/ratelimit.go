package middleware

import (
	"github.com/go-chi/httprate"
	"net/http"
	"time"
)

// RateLimiter защищает от брутфорса и спама
func RateLimiter() func(http.Handler) http.Handler {
	// Ограничение: 60 запросов в минуту с одного IP
	return httprate.LimitByIP(60, 1*time.Minute)
}
