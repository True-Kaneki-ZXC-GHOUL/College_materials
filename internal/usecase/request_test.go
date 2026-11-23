package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/portal/internal/domain"
)

func TestCanReadRequest(t *testing.T) {
	req := &domain.Request{AuthorID: 5, MaterialAuthorID: 9}

	if !canReadRequest(domain.Viewer{UserID: 5, Role: domain.RoleStudent}, req) {
		t.Fatal("студент видит свою заявку")
	}
	if canReadRequest(domain.Viewer{UserID: 6, Role: domain.RoleStudent}, req) {
		t.Fatal("чужую заявку студент не видит")
	}
	if !canReadRequest(domain.Viewer{UserID: 9, Role: domain.RoleTeacher}, req) {
		t.Fatal("преподаватель видит заявки на свой материал")
	}
	if canReadRequest(domain.Viewer{UserID: 8, Role: domain.RoleTeacher}, req) {
		t.Fatal("преподаватель не видит заявки на чужой материал")
	}
	if !canReadRequest(domain.Viewer{UserID: 1, Role: domain.RoleAdmin}, req) {
		t.Fatal("админ видит все заявки")
	}
}

func TestRequestOverdueFlag(t *testing.T) {
	at := time.Now().Add(-6 * 24 * time.Hour)
	req := &domain.Request{Status: domain.TicketStatusNew, OverdueAt: &at}
	req.Overdue = req.Status == domain.TicketStatusNew && req.OverdueAt != nil
	if !req.Overdue {
		t.Fatal("новая заявка с отметкой считается просроченной")
	}

	req.Status = domain.TicketStatusInProgress
	req.Overdue = req.Status == domain.TicketStatusNew && req.OverdueAt != nil
	if req.Overdue {
		t.Fatal("после смены статуса заявка больше не просроченная")
	}
}

func TestNormalizeComment(t *testing.T) {
	if _, err := normalizeComment("   "); err == nil {
		t.Fatal("пустой комментарий не принимается")
	}
	got, err := normalizeComment("  опечатка  ")
	if err != nil || got != "опечатка" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := normalizeComment(strings.Repeat("а", 2001)); err == nil {
		t.Fatal("слишком длинный комментарий не принимается")
	}
}

func TestScopeRequestFilter(t *testing.T) {
	student := domain.Viewer{UserID: 4, Role: domain.RoleStudent}
	f, err := scopeRequestFilter(student, domain.RequestFilter{})
	if err != nil || f.AuthorID == nil || *f.AuthorID != 4 {
		t.Fatalf("студенту должен ограничиваться список своими заявками: %+v %v", f, err)
	}

	teacher := domain.Viewer{UserID: 9, Role: domain.RoleTeacher}
	f, err = scopeRequestFilter(teacher, domain.RequestFilter{})
	if err != nil || f.TeacherID == nil || *f.TeacherID != 9 {
		t.Fatalf("преподавателю только заявки на свои материалы: %+v %v", f, err)
	}
}
