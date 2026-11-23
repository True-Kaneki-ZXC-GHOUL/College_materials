package usecase

import (
	"strings"
	"testing"

	"github.com/yourorg/portal/internal/domain"
)

func TestCanView(t *testing.T) {
	group := int64(7)
	other := int64(8)
	material := &domain.Material{
		AuthorID: 10,
		Status:   domain.MaterialStatusPublished,
		IsPublic: false,
		GroupIDs: []int64{group},
	}

	student := domain.Viewer{UserID: 2, Role: domain.RoleStudent, GroupID: &group}
	outsider := domain.Viewer{UserID: 3, Role: domain.RoleStudent, GroupID: &other}
	noGroup := domain.Viewer{UserID: 4, Role: domain.RoleStudent}
	teacher := domain.Viewer{UserID: 11, Role: domain.RoleTeacher}
	admin := domain.Viewer{UserID: 1, Role: domain.RoleAdmin}

	if !canView(material, student) {
		t.Fatal("студент своей группы должен видеть закрытый опубликованный материал")
	}
	if canView(material, outsider) || canView(material, noGroup) {
		t.Fatal("чужая группа и студент без группы не должны видеть закрытый материал")
	}
	if !canView(material, teacher) || !canView(material, admin) {
		t.Fatal("преподаватель и админ видят закрытые материалы")
	}

	draft := *material
	draft.Status = domain.MaterialStatusDraft
	draft.IsPublic = true
	if canView(&draft, student) {
		t.Fatal("студент не видит черновики")
	}
	if !canView(&draft, teacher) {
		t.Fatal("преподаватель видит чужие черновики")
	}

	public := *material
	public.IsPublic = true
	public.GroupIDs = nil
	if !canView(&public, noGroup) {
		t.Fatal("публичный материал виден студенту без группы")
	}
}

func TestContentChangeIgnoresAccess(t *testing.T) {
	prev := &domain.Material{
		Title:    "A",
		Status:   domain.MaterialStatusPublished,
		IsPublic: false,
		Tags:     []string{"b", "a"},
		GroupIDs: []int64{1},
	}
	next := *prev
	next.Tags = []string{"a", "b"}
	next.Status = domain.MaterialStatusHidden
	next.IsPublic = true
	next.GroupIDs = nil
	if materialContentChanged(prev, &next) {
		t.Fatal("статус, доступ и порядок тегов не должны создавать новую версию")
	}
	next.Title = "B"
	if !materialContentChanged(prev, &next) {
		t.Fatal("смена заголовка — это новая версия")
	}
}

func TestNeedsNewVersion(t *testing.T) {
	if needsNewVersion(domain.MaterialStatusDraft, true) {
		t.Fatal("правка черновика не создаёт версию")
	}
	if needsNewVersion(domain.MaterialStatusPublished, false) {
		t.Fatal("публикация и смена статуса без правки текста не создают версию")
	}
	if !needsNewVersion(domain.MaterialStatusPublished, true) {
		t.Fatal("правка уже опубликованного материала создаёт версию")
	}
}

func TestValidateMaterial(t *testing.T) {
	text := &domain.Material{
		Title:   "Лекция",
		Type:    domain.MaterialTypeText,
		Status:  domain.MaterialStatusPublished,
		Content: "<p>Текст</p><script>alert(1)</script>",
	}
	if err := validateMaterial(text); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text.Content, "script") {
		t.Fatalf("script должен быть вырезан, осталось %q", text.Content)
	}

	link := &domain.Material{
		Title:       "Сайт",
		Type:        domain.MaterialTypeLink,
		Status:      domain.MaterialStatusPublished,
		ExternalURL: "https://go.dev",
	}
	if err := validateMaterial(link); err != nil {
		t.Fatal(err)
	}

	badLink := &domain.Material{
		Title:       "Плохая ссылка",
		Type:        domain.MaterialTypeLink,
		Status:      domain.MaterialStatusPublished,
		ExternalURL: "ftp://files.local/a",
	}
	if err := validateMaterial(badLink); err == nil {
		t.Fatal("ftp-ссылка не должна проходить")
	}

	pdf := &domain.Material{
		Title:  "Методичка",
		Type:   domain.MaterialTypePDF,
		Status: domain.MaterialStatusPublished,
	}
	if err := validateMaterial(pdf); err == nil {
		t.Fatal("нельзя опубликовать pdf без файла")
	}
	pdf.Status = domain.MaterialStatusDraft
	if err := validateMaterial(pdf); err != nil {
		t.Fatal(err)
	}

	docx := &domain.Material{Title: "Word", Type: "docx", Status: domain.MaterialStatusDraft}
	if err := validateMaterial(docx); err == nil {
		t.Fatal("docx больше не поддерживается")
	}

	both := &domain.Material{
		Title:    "Всем и не всем",
		Type:     domain.MaterialTypeText,
		Status:   domain.MaterialStatusDraft,
		Content:  "<p>x</p>",
		IsPublic: true,
		GroupIDs: []int64{1},
	}
	if err := validateMaterial(both); err == nil {
		t.Fatal("публичный материал нельзя ограничить группой")
	}
}
