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
	groups  domain.GroupRepository
}

func NewInviteUsecase(invites domain.InviteRepository, groups domain.GroupRepository) *InviteUsecase {
	return &InviteUsecase{invites: invites, groups: groups}
}

// Create — 4.2. Срок — 3 дня по умолчанию.
func (uc *InviteUsecase) Create(ctx context.Context, groupID, createdBy int64) (*domain.InviteCode, error) {
	if groupID <= 0 {
		return nil, domain.ErrValidation
	}
	// 404, если группы нет — вместо FK-ошибки на INSERT.
	if _, err := uc.groups.GetByID(ctx, groupID); err != nil {
		return nil, err
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
