package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/pkg/storage"
)

type MaterialUsecase struct {
	repo       domain.MaterialRepository
	storage    storage.Storage
	presignTTL time.Duration
}

func NewMaterialUsecase(repo domain.MaterialRepository, storage storage.Storage, presignTTL time.Duration) *MaterialUsecase {
	return &MaterialUsecase{repo: repo, storage: storage, presignTTL: presignTTL}
}

type CreateMaterialInput struct {
	Title       string
	Description string
	Type        domain.MaterialType
	Status      domain.MaterialStatus
	SubjectID   int64
	Tags        []string
	IsPublic    bool
	FileKey     string
	Content     string
	ExternalURL string
	GroupIDs    []int64
}

func (uc *MaterialUsecase) Create(ctx context.Context, authorID int64, in CreateMaterialInput) (*domain.Material, error) {
	if in.Status == "" {
		in.Status = domain.MaterialStatusDraft
	}
	m := &domain.Material{
		Title:       in.Title,
		Description: strings.TrimSpace(in.Description),
		Type:        in.Type,
		Status:      in.Status,
		SubjectID:   in.SubjectID,
		AuthorID:    authorID,
		IsPublic:    in.IsPublic,
		FileKey:     strings.TrimSpace(in.FileKey),
		Content:     in.Content,
		ExternalURL: in.ExternalURL,
		GroupIDs:    in.GroupIDs,
		IsActive:    true,
	}
	tags, err := normalizeTags(in.Tags)
	if err != nil {
		return nil, err
	}
	m.Tags = tags
	if m.IsPublic {
		if len(uniqueIDs(in.GroupIDs)) > 0 {
			return nil, invalid("публичный материал нельзя ограничивать группами")
		}
		m.GroupIDs = []int64{}
	}
	if err := validateMaterial(m); err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateMaterialInput struct {
	Title       *string
	Description *string
	Type        *domain.MaterialType
	Status      *domain.MaterialStatus
	SubjectID   *int64
	Tags        *[]string
	IsPublic    *bool
	GroupIDs    *[]int64
	FileKey     *string
	Content     *string
	ExternalURL *string
}

func (uc *MaterialUsecase) Update(ctx context.Context, v domain.Viewer, id int64, in UpdateMaterialInput) (*domain.Material, error) {
	current, err := uc.editable(ctx, v, id)
	if err != nil {
		return nil, err
	}

	next := *current
	next.Tags = append([]string(nil), current.Tags...)
	next.GroupIDs = append([]int64(nil), current.GroupIDs...)

	if in.Title != nil {
		next.Title = *in.Title
	}
	if in.Description != nil {
		next.Description = strings.TrimSpace(*in.Description)
	}
	if in.Type != nil {
		next.Type = *in.Type
	}
	if in.Status != nil {
		next.Status = *in.Status
	}
	if in.SubjectID != nil {
		next.SubjectID = *in.SubjectID
	}
	if in.Tags != nil {
		tags, err := normalizeTags(*in.Tags)
		if err != nil {
			return nil, err
		}
		next.Tags = tags
	}
	if in.IsPublic != nil {
		next.IsPublic = *in.IsPublic
	}
	if in.GroupIDs != nil {
		next.GroupIDs = *in.GroupIDs
	}
	if in.FileKey != nil {
		next.FileKey = strings.TrimSpace(*in.FileKey)
	}
	if in.Content != nil {
		next.Content = *in.Content
	}
	if in.ExternalURL != nil {
		next.ExternalURL = *in.ExternalURL
	}
	// Публичный материал виден всем, список групп для него не нужен.
	if in.IsPublic != nil && *in.IsPublic && in.GroupIDs == nil {
		next.GroupIDs = []int64{}
	}

	if err := validateMaterial(&next); err != nil {
		return nil, err
	}

	// Новая версия только если правят уже опубликованный материал.
	// Смена статуса, видимости и списка групп историю не плодит.
	if needsNewVersion(current.Status, materialContentChanged(current, &next)) {
		if err := uc.repo.CreateVersion(ctx, current, &next); err != nil {
			return nil, err
		}
		return &next, nil
	}
	if err := uc.repo.Update(ctx, &next); err != nil {
		return nil, err
	}
	return &next, nil
}

func (uc *MaterialUsecase) ChangeStatus(ctx context.Context, v domain.Viewer, id int64, status domain.MaterialStatus) (*domain.Material, error) {
	return uc.Update(ctx, v, id, UpdateMaterialInput{Status: &status})
}

func (uc *MaterialUsecase) Delete(ctx context.Context, v domain.Viewer, id int64) error {
	if _, err := uc.editable(ctx, v, id); err != nil {
		return err
	}
	return uc.repo.SoftDelete(ctx, id)
}

func (uc *MaterialUsecase) Restore(ctx context.Context, v domain.Viewer, id int64) error {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if v.Role != domain.RoleAdmin && m.AuthorID != v.UserID {
		return domain.ErrForbidden
	}
	if m.DeletedAt == nil {
		return invalid("материал не удалён")
	}
	return uc.repo.Restore(ctx, id)
}

func (uc *MaterialUsecase) GetByID(ctx context.Context, v domain.Viewer, id int64) (*domain.Material, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	// Старые версии студенту не отдаём: в каталоге только активная.
	if !m.IsActive && v.Role == domain.RoleStudent {
		return nil, domain.ErrNotFound
	}
	if !canView(m, v) {
		return nil, domain.ErrNotFound
	}
	return m, nil
}

func (uc *MaterialUsecase) List(ctx context.Context, v domain.Viewer, f domain.MaterialFilter) ([]*domain.Material, int, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	tags, err := normalizeTags(f.Tags)
	if err != nil {
		return nil, 0, err
	}
	f.Tags = tags
	if f.Type != nil && !validType(*f.Type) {
		return nil, 0, invalid("type должен быть pdf, link или text")
	}
	if f.Status != nil && !validStatus(*f.Status) {
		return nil, 0, invalid("status должен быть draft, published или hidden")
	}
	// Студент не выбирает чужую группу и не открывает корзину.
	if v.Role == domain.RoleStudent {
		f.GroupID = nil
		f.IncludeDeleted = false
	} else if f.IncludeDeleted && v.Role == domain.RoleTeacher {
		f.AuthorID = &v.UserID
	}
	f.Viewer = v
	return uc.repo.List(ctx, f)
}

func (uc *MaterialUsecase) ListVersions(ctx context.Context, v domain.Viewer, id int64) ([]*domain.Material, error) {
	if v.Role == domain.RoleStudent {
		return nil, domain.ErrForbidden
	}
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return uc.repo.ListVersions(ctx, id)
}

func (uc *MaterialUsecase) UploadFile(
	ctx context.Context, v domain.Viewer, id int64,
	mtype domain.MaterialType, r io.Reader, size int64,
) (string, int64, error) {
	if uc.storage == nil {
		return "", 0, fmt.Errorf("%w: хранилище файлов не настроено", domain.ErrUnavailable)
	}
	m, err := uc.editable(ctx, v, id)
	if err != nil {
		return "", 0, err
	}
	if m.Type != domain.MaterialTypePDF || mtype != domain.MaterialTypePDF {
		return "", 0, invalid("файл можно прикрепить только к материалу типа pdf")
	}

	key := fmt.Sprintf("materials/%d/%s.pdf", id, randomHex(16))
	if err := uc.storage.Upload(ctx, key, r, size, "application/pdf"); err != nil {
		return "", 0, err
	}

	if m.Status == domain.MaterialStatusPublished {
		next := *m
		next.Tags = append([]string(nil), m.Tags...)
		next.GroupIDs = append([]int64(nil), m.GroupIDs...)
		next.FileKey = key
		if err := uc.repo.CreateVersion(ctx, m, &next); err != nil {
			_ = uc.storage.Delete(ctx, key)
			return "", 0, err
		}
		return key, next.ID, nil
	}

	if err := uc.repo.UpdateFileKey(ctx, id, key); err != nil {
		_ = uc.storage.Delete(ctx, key)
		return "", 0, err
	}
	if m.FileKey != "" && m.FileKey != key {
		_ = uc.storage.Delete(ctx, m.FileKey)
	}
	return key, id, nil
}

func (uc *MaterialUsecase) GetDownloadURL(ctx context.Context, v domain.Viewer, id int64) (string, error) {
	if uc.storage == nil {
		return "", fmt.Errorf("%w: хранилище файлов не настроено", domain.ErrUnavailable)
	}
	m, err := uc.GetByID(ctx, v, id)
	if err != nil {
		return "", err
	}
	if m.Type != domain.MaterialTypePDF || m.FileKey == "" {
		return "", domain.ErrNotFound
	}
	return uc.storage.PresignedGetURL(ctx, m.FileKey, uc.presignTTL)
}

func (uc *MaterialUsecase) editable(ctx context.Context, v domain.Viewer, id int64) (*domain.Material, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.DeletedAt != nil || !m.IsActive {
		return nil, domain.ErrNotFound
	}
	if v.Role != domain.RoleAdmin && m.AuthorID != v.UserID {
		return nil, domain.ErrForbidden
	}
	return m, nil
}

// canView: преподаватель и админ читают всё.
// Студент видит только опубликованное публичное или материал своей группы.
func canView(m *domain.Material, v domain.Viewer) bool {
	switch v.Role {
	case domain.RoleAdmin, domain.RoleTeacher:
		return true
	case domain.RoleStudent:
		if m.Status != domain.MaterialStatusPublished {
			return false
		}
		if m.IsPublic {
			return true
		}
		if v.GroupID == nil {
			return false
		}
		for _, g := range m.GroupIDs {
			if g == *v.GroupID {
				return true
			}
		}
	}
	return false
}

func needsNewVersion(current domain.MaterialStatus, contentChanged bool) bool {
	return current == domain.MaterialStatusPublished && contentChanged
}

func materialContentChanged(prev, next *domain.Material) bool {
	if prev.Title != next.Title || prev.Description != next.Description || prev.Type != next.Type {
		return true
	}
	if prev.Content != next.Content || prev.ExternalURL != next.ExternalURL || prev.FileKey != next.FileKey {
		return true
	}
	if prev.SubjectID != next.SubjectID || !sameStringSet(prev.Tags, next.Tags) {
		return true
	}
	return false
}

func validateMaterial(m *domain.Material) error {
	m.Title = strings.TrimSpace(m.Title)
	if m.Title == "" {
		return invalid("нужен заголовок")
	}
	if len([]rune(m.Title)) > 255 {
		return invalid("заголовок длиннее 255 символов")
	}
	if len([]rune(m.Description)) > 5000 {
		return invalid("описание длиннее 5000 символов")
	}
	if !validType(m.Type) {
		return invalid("type должен быть pdf, link или text")
	}
	if !validStatus(m.Status) {
		return invalid("status должен быть draft, published или hidden")
	}

	m.Content = sanitizeContent(m.Content)
	if len(m.Content) > 500_000 {
		return invalid("текст материала слишком длинный")
	}
	m.ExternalURL = strings.TrimSpace(m.ExternalURL)
	if m.ExternalURL != "" && !validExternalURL(m.ExternalURL) {
		return invalid("external_url должен быть http или https ссылкой")
	}
	if len(m.ExternalURL) > 2000 {
		return invalid("ссылка слишком длинная")
	}

	m.GroupIDs = uniqueIDs(m.GroupIDs)
	if m.IsPublic && len(m.GroupIDs) > 0 {
		return invalid("публичный материал нельзя ограничивать группами")
	}

	publishing := m.Status == domain.MaterialStatusPublished
	switch m.Type {
	case domain.MaterialTypeLink:
		if publishing && !validExternalURL(m.ExternalURL) {
			return invalid("для публикации ссылки нужен external_url")
		}
	case domain.MaterialTypeText:
		if publishing && strings.TrimSpace(m.Content) == "" {
			return invalid("для публикации текста нужно поле content")
		}
	case domain.MaterialTypePDF:
		if publishing && strings.TrimSpace(m.FileKey) == "" {
			return invalid("сначала загрузите PDF, потом публикуйте")
		}
	}
	return nil
}

func validType(t domain.MaterialType) bool {
	switch t {
	case domain.MaterialTypePDF, domain.MaterialTypeLink, domain.MaterialTypeText:
		return true
	}
	return false
}

func validStatus(s domain.MaterialStatus) bool {
	switch s {
	case domain.MaterialStatusDraft, domain.MaterialStatusPublished, domain.MaterialStatusHidden:
		return true
	}
	return false
}

func validExternalURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func normalizeTags(in []string) ([]string, error) {
	if len(in) == 0 {
		return []string{}, nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if len([]rune(t)) > 50 {
			return nil, invalid("тег длиннее 50 символов")
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	if len(out) > 20 {
		return nil, invalid("не больше 20 тегов")
	}
	return out, nil
}

func uniqueIDs(in []int64) []int64 {
	if len(in) == 0 {
		return []int64{}
	}
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, id := range in {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	n := map[string]int{}
	for _, s := range a {
		n[s]++
	}
	for _, s := range b {
		if n[s] == 0 {
			return false
		}
		n[s]--
	}
	return true
}

var (
	reDangerousTag = regexp.MustCompile(`(?is)<\s*(script|iframe|object|embed)\b[^>]*>.*?<\s*/\s*(script|iframe|object|embed)\s*>`)
	reDangerousOne = regexp.MustCompile(`(?is)<\s*(script|iframe|object|embed)\b[^>]*/?>`)
	reEventAttr    = regexp.MustCompile(`(?i)\son[a-z0-9_-]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	reJSURL        = regexp.MustCompile(`(?i)javascript\s*:`)
)

// sanitizeContent убирает только явно опасные куски.
// Обычный HTML редактора (абзацы, списки, ссылки) сохраняется как есть.
func sanitizeContent(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = reDangerousTag.ReplaceAllString(s, "")
	s = reDangerousOne.ReplaceAllString(s, "")
	s = reEventAttr.ReplaceAllString(s, "")
	s = reJSURL.ReplaceAllString(s, "")
	return s
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
