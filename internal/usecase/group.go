package usecase

import (
	"context"

	"github.com/yourorg/portal/internal/domain"
)

type GroupUsecase struct {
	repo domain.GroupRepository
}

func NewGroupUsecase(repo domain.GroupRepository) *GroupUsecase {
	return &GroupUsecase{repo: repo}
}

func (uc *GroupUsecase) Create(ctx context.Context, g *domain.Group) error {
	if g.Name == "" {
		return domain.ErrValidation
	}
	return uc.repo.Create(ctx, g)
}

func (uc *GroupUsecase) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *GroupUsecase) Update(ctx context.Context, g *domain.Group) error {
	if g.Name == "" {
		return domain.ErrValidation
	}
	return uc.repo.Update(ctx, g)
}

func (uc *GroupUsecase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *GroupUsecase) List(ctx context.Context) ([]*domain.Group, error) {
	return uc.repo.List(ctx)
}
