package usecase

import (
	"context"

	"github.com/yourorg/portal/internal/domain"
)

type UserUsecase struct {
	users domain.UserRepository
}

func NewUserUsecase(users domain.UserRepository) *UserUsecase {
	return &UserUsecase{users: users}
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return uc.users.GetByID(ctx, id)
}

func (uc *UserUsecase) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.users.List(ctx, limit, offset)
}