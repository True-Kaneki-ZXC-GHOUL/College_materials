package usecase

import (
	"fmt"
	"strings"

	"github.com/yourorg/portal/internal/domain"
)

func invalid(msg string) error {
	return fmt.Errorf("%w: %s", domain.ErrValidation, msg)
}

func normalizeEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || len(s) > 255 || !strings.Contains(s, "@") || strings.ContainsAny(s, " \t\r\n") {
		return "", invalid("некорректный email")
	}
	return s, nil
}
