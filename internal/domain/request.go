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
	ID               int64        `json:"id"`
	MaterialID       int64        `json:"material_id"`
	MaterialTitle    string       `json:"material_title"`
	MaterialAuthorID int64        `json:"-"`
	AuthorID         int64        `json:"author_id"`
	Comment          string       `json:"comment"`
	Status           TicketStatus `json:"status"`
	Overdue          bool         `json:"overdue"`
	OverdueAt        *time.Time   `json:"overdue_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type RequestFilter struct {
	Status     *TicketStatus
	MaterialID *int64
	AuthorID   *int64
	TeacherID  *int64
	Overdue    bool
	Page       int
	PageSize   int
}
