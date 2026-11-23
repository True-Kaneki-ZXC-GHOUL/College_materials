package domain

// Viewer — контекст запрашивающего для проверки прав на материалы.
// GroupID нужен только для студентов — nil, если группы нет.
type Viewer struct {
	UserID  int64
	Role    Role
	GroupID *int64
}
