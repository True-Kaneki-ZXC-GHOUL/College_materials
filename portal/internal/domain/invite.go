package domain

import "time"

type InviteCode struct {
	ID        int64
	Code      string
	GroupID   int64
	CreatedBy int64
	ExpiresAt time.Time
	UsedAt    *time.Time
	UsedBy    *int64
}