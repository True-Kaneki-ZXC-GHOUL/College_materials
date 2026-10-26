package repository

import (
	"context"
	"sync"

	"github.com/yourorg/portal/internal/domain"
)

type InMemoryUserRepo struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]*domain.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		nextID: 1,
		items:  make(map[int64]*domain.User),
	}
}

func (r *InMemoryUserRepo) Create(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u.ID = r.nextID
	r.nextID++
	r.items[u.ID] = u
	return nil
}

func (r *InMemoryUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (r *InMemoryUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.items {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryUserRepo) Update(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[u.ID]; !ok {
		return domain.ErrNotFound
	}
	r.items[u.ID] = u
	return nil
}

func (r *InMemoryUserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]*domain.User, 0, len(r.items))
	for _, u := range r.items {
		res = append(res, u)
	}
	return res, nil
}