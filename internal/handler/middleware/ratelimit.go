package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientLimit struct {
	count   int
	started time.Time
}

var (
	limitsMu sync.Mutex
	limits   = make(map[string]clientLimit)
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		limitsMu.Lock()

		now := time.Now()
		limit, exists := limits[ip]

		if !exists || now.Sub(limit.started) >= time.Minute {
			limit = clientLimit{count: 0, started: now}
		}

		limit.count++
		limits[ip] = limit
		allowed := limit.count <= 5

		limitsMu.Unlock()

		if !allowed {
			http.Error(w, "Слишком много запросов", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
