package middleware

import (
	"net/http"
	"time"
	"github.com/go-chi/httprate"
)

// RateLimiter защищает от брутфорса и спама
func RateLimiter() func(http.Handler) http.Handler {
	// Ограничение: 60 запросов в минуту с одного IP
	return httprate.LimitByIP(60, 1*time.Minute)
}