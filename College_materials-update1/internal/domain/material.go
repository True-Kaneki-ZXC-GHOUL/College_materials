package domain

import "time"

type MaterialType string
type MaterialStatus string

const (
	MaterialTypePDF  MaterialType = "pdf"
	MaterialTypeDOCX MaterialType = "docx"
	MaterialTypeLink MaterialType = "link"

	MaterialStatusDraft     MaterialStatus = "draft"
	MaterialStatusPublished MaterialStatus = "published"
	MaterialStatusHidden    MaterialStatus = "hidden"
)

type Material struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        MaterialType   `json:"type"`
	Status      MaterialStatus `json:"status"`
	SubjectID   int64          `json:"subject_id"`
	AuthorID    int64          `json:"author_id"`
	Tags        []string       `json:"tags"`
	IsPublic    bool           `json:"is_public"`
	GroupIDs    []int64        `json:"group_ids"`
	FileKey     string         `json:"file_key"`
	ReadTimeEst int            `json:"read_time_est"`
	ViewsCount  int64          `json:"views_count"`
	IsActive    bool           `json:"is_active"`
	ParentID    *int64         `json:"parent_id"`
	DeletedAt   *time.Time     `json:"deleted_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
type MaterialFilter struct {
	Query     string
	SubjectID *int64
	Status    *MaterialStatus
	Tags      []string
	AuthorID  *int64
	OnlyMine  bool
	Page      int
	PageSize  int
	SortBy    string
}
