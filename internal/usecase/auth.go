package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/pkg/password"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type AuthUsecase struct {
	users              domain.UserRepository
	invites            domain.InviteRepository
	sessions           domain.SessionRepository
	jwtSecret          string
	accessTTL          time.Duration
	refreshTTL         time.Duration
	sessionIdleTimeout time.Duration
}

func NewAuthUsecase(
	users domain.UserRepository,
	invites domain.InviteRepository,
	sessions domain.SessionRepository,
	jwtSecret string,
	accessTTL, refreshTTL, sessionIdleTimeout time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		users:              users,
		invites:            invites,
		sessions:           sessions,
		jwtSecret:          jwtSecret,
		accessTTL:          accessTTL,
		refreshTTL:         refreshTTL,
		sessionIdleTimeout: sessionIdleTimeout,
	}
}

func (uc *AuthUsecase) Login(ctx context.Context, email, rawPassword, ip, ua string) (*TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	ok, err := password.Verify(rawPassword, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrUnauthorized
	}
	if user.IsBlocked {
		return nil, domain.ErrForbidden
	}
	return uc.startSession(ctx, user, ip, ua)
}

func (uc *AuthUsecase) Register(ctx context.Context, email, rawPassword, fullName, code, ip, ua string) (*TokenPair, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	fullName = strings.TrimSpace(fullName)
	if len(rawPassword) < 8 || fullName == "" || strings.TrimSpace(code) == "" {
		return nil, invalid("нужны email, имя, пароль от 8 символов и инвайт-код")
	}
	if len([]rune(fullName)) > 255 {
		return nil, invalid("имя длиннее 255 символов")
	}

	inv, err := uc.invites.GetByCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, invalid("инвайт-код не найден")
	}
	if inv.UsedAt != nil {
		return nil, invalid("инвайт-код уже использован")
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, invalid("срок инвайт-кода истёк")
	}

	hash, err := password.Hash(rawPassword)
	if err != nil {
		return nil, err
	}

	groupID := inv.GroupID
	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         domain.RoleStudent,
		GroupID:      &groupID,
	}
	if err := uc.users.RegisterStudent(ctx, user, inv.ID); err != nil {
		return nil, err
	}
	return uc.startSession(ctx, user, ip, ua)
}

func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, domain.ErrUnauthorized
	}
	sess, err := uc.sessions.GetByRefreshHash(ctx, hashRefresh(refreshToken))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	now := time.Now()
	if now.After(sess.ExpiresAt) || now.Sub(sess.LastActivityAt) > uc.sessionIdleTimeout {
		_ = uc.sessions.Delete(ctx, sess.ID)
		return nil, domain.ErrUnauthorized
	}

	user, err := uc.users.GetByID(ctx, sess.UserID)
	if err != nil || user.IsBlocked {
		_ = uc.sessions.Delete(ctx, sess.ID)
		return nil, domain.ErrUnauthorized
	}

	newRefresh, newHash, err := generateRefresh()
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.UpdateRefresh(ctx, sess.ID, newHash, now, now.Add(uc.refreshTTL)); err != nil {
		return nil, err
	}

	access, err := uc.signAccess(user)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    int(uc.accessTTL.Seconds()),
	}, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	sess, err := uc.sessions.GetByRefreshHash(ctx, hashRefresh(refreshToken))
	if err != nil {
		return nil
	}
	return uc.sessions.Delete(ctx, sess.ID)
}

func (uc *AuthUsecase) VerifyAccessToken(tokenString string) (domain.Viewer, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return domain.Viewer{}, domain.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return domain.Viewer{}, domain.ErrUnauthorized
	}
	if t, _ := claims["type"].(string); t != "access" {
		return domain.Viewer{}, domain.ErrUnauthorized
	}
	uidF, ok := claims["user_id"].(float64)
	if !ok {
		return domain.Viewer{}, domain.ErrUnauthorized
	}
	roleS, ok := claims["role"].(string)
	if !ok || !validRole(domain.Role(roleS)) {
		return domain.Viewer{}, domain.ErrUnauthorized
	}
	var gid *int64
	if gf, ok := claims["group_id"].(float64); ok && gf > 0 {
		v := int64(gf)
		gid = &v
	}
	return domain.Viewer{UserID: int64(uidF), Role: domain.Role(roleS), GroupID: gid}, nil
}

func validRole(r domain.Role) bool {
	switch r {
	case domain.RoleAdmin, domain.RoleTeacher, domain.RoleStudent:
		return true
	}
	return false
}

func (uc *AuthUsecase) startSession(ctx context.Context, user *domain.User, ip, ua string) (*TokenPair, error) {
	refresh, refreshHash, err := generateRefresh()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sess := &domain.Session{
		ID:               newUUIDv4(),
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		IPAddress:        ip,
		UserAgent:        ua,
		LastActivityAt:   now,
		ExpiresAt:        now.Add(uc.refreshTTL),
	}
	if err := uc.sessions.Create(ctx, sess); err != nil {
		return nil, err
	}
	access, err := uc.signAccess(user)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(uc.accessTTL.Seconds()),
	}, nil
}

func (uc *AuthUsecase) signAccess(u *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"role":    string(u.Role),
		"type":    "access",
		"iat":     now.Unix(),
		"exp":     now.Add(uc.accessTTL).Unix(),
	}
	if u.GroupID != nil {
		claims["group_id"] = *u.GroupID
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(uc.jwtSecret))
}

func generateRefresh() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func hashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newUUIDv4() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
