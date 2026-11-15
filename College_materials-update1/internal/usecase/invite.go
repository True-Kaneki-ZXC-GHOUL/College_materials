package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/yourorg/portal/internal/domain"
)

type InviteUsecase struct {
	invites domain.InviteRepository
}

func NewInviteUsecase(invites domain.InviteRepository) *InviteUsecase {
	return &InviteUsecase{invites: invites}
}

// Create — минимальная версия задачи 4.2. Срок — 3 дня по умолчанию.
func (uc *InviteUsecase) Create(ctx context.Context, groupID, createdBy int64) (*domain.InviteCode, error) {
	if groupID <= 0 {
		return nil, domain.ErrValidation
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := base64.RawURLEncoding.EncodeToString(b)

	inv := &domain.InviteCode{
		Code:      code,
		GroupID:   groupID,
		CreatedBy: createdBy,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}
	if err := uc.invites.Create(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}