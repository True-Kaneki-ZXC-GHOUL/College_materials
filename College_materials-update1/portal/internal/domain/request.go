package domain

import "time"

type TicketStatus string

const (
	TicketStatusNew        TicketStatus = "new"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusDone       TicketStatus = "done"
	TicketStatusRejected   TicketStatus = "rejected"
)

type Request struct {
	ID         int64
	MaterialID int64
	AuthorID   int64
	Comment    string
	Status     TicketStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}