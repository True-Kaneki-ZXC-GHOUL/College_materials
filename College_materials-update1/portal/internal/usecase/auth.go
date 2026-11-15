package usecase

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/portal/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthUsecase struct {
	users     domain.UserRepository
	jwtSecret string
}

func NewAuthUsecase(users domain.UserRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{users: users, jwtSecret: jwtSecret}
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(uc.jwtSecret))
}
