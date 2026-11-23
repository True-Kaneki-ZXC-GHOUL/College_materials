package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/portal/internal/domain"
	mw "github.com/yourorg/portal/internal/handler/middleware"
)

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// currentUser достаёт id и роль из контекста (заполняет middleware.Auth).
func currentUser(r *http.Request) (int64, domain.Role, bool) {
	uid, ok1 := r.Context().Value(mw.UserIDKey).(int64)
	role, ok2 := r.Context().Value(mw.UserRoleKey).(domain.Role)
	return uid, role, ok1 && ok2
}

func currentViewer(r *http.Request) (domain.Viewer, bool) {
	uid, ok1 := r.Context().Value(mw.UserIDKey).(int64)
	role, ok2 := r.Context().Value(mw.UserRoleKey).(domain.Role)
	if !ok1 || !ok2 {
		return domain.Viewer{}, false
	}
	gid, _ := r.Context().Value(mw.UserGroupKey).(*int64)
	return domain.Viewer{UserID: uid, Role: role, GroupID: gid}, true
}
