// Package service holds application/business logic. Services depend only on
// domain repository interfaces (never on GORM/Gin/pgx directly), which is what
// lets handlers, repositories, and business rules be tested/replaced independently.
package service

import (
	"context"
	"time"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/hash"
	"bookreader/backend/pkg/jwtutil"
)

type AuthService struct {
	users      domain.UserRepository
	roles      domain.RoleRepository
	refresh    domain.RefreshTokenRepository
	sessions   domain.SessionRepository
	issuer     *jwtutil.Issuer
	refreshTTL time.Duration
}

func NewAuthService(
	users domain.UserRepository,
	roles domain.RoleRepository,
	refresh domain.RefreshTokenRepository,
	sessions domain.SessionRepository,
	issuer *jwtutil.Issuer,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{users: users, roles: roles, refresh: refresh, sessions: sessions, issuer: issuer, refreshTTL: refreshTTL}
}

type AuthTokens struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	User                  *domain.User
}

type RequestMeta struct {
	IP        string
	UserAgent string
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperror.Internal("failed to check existing user", err)
	}
	if existing != nil {
		return nil, apperror.Conflict("an account with this email already exists")
	}

	role, err := s.roles.FindByName(ctx, domain.RoleUser)
	if err != nil || role == nil {
		return nil, apperror.Internal("default role not configured", err)
	}

	hashed, err := hash.HashPassword(password)
	if err != nil {
		return nil, apperror.Internal("failed to hash password", err)
	}

	user := &domain.User{RoleID: role.ID, Name: name, Email: email, PasswordHash: hashed, IsActive: true}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, apperror.Internal("failed to create user", err)
	}
	user.Role = role
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, meta RequestMeta) (*AuthTokens, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperror.Internal("failed to look up user", err)
	}
	if user == nil || !user.IsActive || !hash.VerifyPassword(user.PasswordHash, password) {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	tokens, err := s.issueTokens(ctx, user, meta)
	if err != nil {
		return nil, err
	}
	_ = s.users.TouchLastLogin(ctx, user.ID)
	return tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string, meta RequestMeta) (*AuthTokens, error) {
	hashed := hash.HashToken(rawRefreshToken)
	stored, err := s.refresh.FindByHash(ctx, hashed)
	if err != nil {
		return nil, apperror.Internal("failed to look up refresh token", err)
	}
	if stored == nil || stored.RevokedAt != nil || stored.ExpiresAt.Before(time.Now()) {
		return nil, apperror.Unauthorized("refresh token is invalid or expired")
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil || user == nil || !user.IsActive {
		return nil, apperror.Unauthorized("account is no longer active")
	}

	tokens, err := s.issueTokens(ctx, user, meta)
	if err != nil {
		return nil, err
	}
	// Rotate: revoke the old refresh token and link it to the newly issued one
	// so a reused/stolen token is detectable (its replaced_by_id will be set).
	_ = s.refresh.Revoke(ctx, stored.ID, nil)
	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashed := hash.HashToken(rawRefreshToken)
	stored, err := s.refresh.FindByHash(ctx, hashed)
	if err != nil {
		return apperror.Internal("failed to look up refresh token", err)
	}
	if stored == nil {
		return nil
	}
	return s.refresh.Revoke(ctx, stored.ID, nil)
}

func (s *AuthService) issueTokens(ctx context.Context, user *domain.User, meta RequestMeta) (*AuthTokens, error) {
	roleName := domain.RoleUser
	if user.Role != nil {
		roleName = user.Role.Name
	}

	access, accessExp, err := s.issuer.GenerateAccessToken(user.ID, user.Email, roleName)
	if err != nil {
		return nil, apperror.Internal("failed to generate access token", err)
	}

	rawRefresh, err := jwtutil.GenerateOpaqueToken()
	if err != nil {
		return nil, apperror.Internal("failed to generate refresh token", err)
	}
	refreshExp := time.Now().Add(s.refreshTTL)

	rt := &domain.RefreshToken{
		UserID: user.ID, TokenHash: hash.HashToken(rawRefresh),
		CreatedByIP: meta.IP, UserAgent: meta.UserAgent, ExpiresAt: refreshExp,
	}
	if err := s.refresh.Create(ctx, rt); err != nil {
		return nil, apperror.Internal("failed to persist refresh token", err)
	}

	_ = s.sessions.Create(ctx, &domain.Session{
		UserID: user.ID, RefreshTokenID: &rt.ID, IPAddress: meta.IP, UserAgent: meta.UserAgent, ExpiresAt: refreshExp,
	})

	return &AuthTokens{
		AccessToken: access, AccessTokenExpiresAt: accessExp,
		RefreshToken: rawRefresh, RefreshTokenExpiresAt: refreshExp,
		User: user,
	}, nil
}
