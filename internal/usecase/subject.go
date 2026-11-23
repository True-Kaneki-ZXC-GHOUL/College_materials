package usecase

import (
	"context"

	"github.com/yourorg/portal/internal/domain"
)

type SubjectUsecase struct {
	repo domain.SubjectRepository
}

func NewSubjectUsecase(repo domain.SubjectRepository) *SubjectUsecase {
	return &SubjectUsecase{repo: repo}
}

func (uc *SubjectUsecase) Create(ctx context.Context, s *domain.Subject) error {
	if s.Name == "" {
		return domain.ErrValidation
	}
	return uc.repo.Create(ctx, s)
}

func (uc *SubjectUsecase) GetByID(ctx context.Context, id int64) (*domain.Subject, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *SubjectUsecase) Update(ctx context.Context, s *domain.Subject) error {
	if s.Name == "" {
		return domain.ErrValidation
	}
	return uc.repo.Update(ctx, s)
}

func (uc *SubjectUsecase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *SubjectUsecase) List(ctx context.Context) ([]*domain.Subject, error) {
	return uc.repo.List(ctx)
}
