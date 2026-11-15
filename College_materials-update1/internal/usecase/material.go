package usecase

import (
	"context"
	"github.com/yourorg/portal/internal/domain"
)

type MaterialUsecase struct {
	repo domain.MaterialRepository
}

func NewMaterialUsecase(repo domain.MaterialRepository) *MaterialUsecase {
	return &MaterialUsecase{repo: repo}
}

func (uc *MaterialUsecase) Create(ctx context.Context, m *domain.Material) error {
	return uc.repo.Create(ctx, m)
}

func (uc *MaterialUsecase) List(ctx context.Context, f domain.MaterialFilter) ([]*domain.Material, int, error) {
	return uc.repo.List(ctx, f)
}

func (uc *MaterialUsecase) GetByID(ctx context.Context, id int64) (*domain.Material, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *MaterialUsecase) Delete(ctx context.Context, id int64) error {
	return uc.repo.SoftDelete(ctx, id)
}
