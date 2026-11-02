package domain

import "time"

type Session struct {
	ID               string
	UserID           int64
	RefreshTokenHash string
	IPAddress        string
	UserAgent        string
	LastActivityAt   time.Time
	ExpiresAt        time.Time
	CreatedAt        time.Time
}