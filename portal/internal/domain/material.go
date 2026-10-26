package domain

import "time"

type MaterialType   string
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
	ID          int64
	Title       string
	Description string
	Type        MaterialType
	Status      MaterialStatus
	SubjectID   int64
	AuthorID    int64
	Tags        []string
	IsPublic    bool
	GroupIDs    []int64
	FileKey     string
	ReadTimeEst int
	ViewsCount  int64
	IsActive    bool
	ParentID    *int64
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
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