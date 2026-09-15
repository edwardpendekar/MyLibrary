package service_test

import (
	"context"
	"testing"
	"time"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/hash"
	"bookreader/backend/pkg/jwtutil"
)

// --- minimal in-memory fakes conforming to the domain interfaces ---
// No mocking framework needed: the domain package's small, explicit
// interfaces are exactly what makes this kind of fake trivial to write.

type fakeUserRepo struct {
	byEmail map[string]*domain.User
	nextID  int64
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{byEmail: map[string]*domain.User{}} }

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	f.nextID++
	u.ID = f.nextID
	f.byEmail[u.Email] = u
	return nil
}
func (f *fakeUserRepo) FindByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return f.byEmail[email], nil
}
func (f *fakeUserRepo) Update(_ context.Context, u *domain.User) error {
	f.byEmail[u.Email] = u
	return nil
}
func (f *fakeUserRepo) SoftDelete(context.Context, int64) error { return nil }
func (f *fakeUserRepo) List(context.Context, string, int) (*domain.ListResult[domain.User], error) {
	return &domain.ListResult[domain.User]{}, nil
}
func (f *fakeUserRepo) TouchLastLogin(context.Context, int64) error { return nil }

type fakeRoleRepo struct{ roles map[string]*domain.Role }

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{roles: map[string]*domain.Role{
		domain.RoleUser:  {ID: 1, Name: domain.RoleUser},
		domain.RoleAdmin: {ID: 2, Name: domain.RoleAdmin},
	}}
}

func (f *fakeRoleRepo) FindByName(_ context.Context, name string) (*domain.Role, error) {
	return f.roles[name], nil
}
func (f *fakeRoleRepo) List(context.Context) ([]domain.Role, error) { return nil, nil }

type fakeRefreshTokenRepo struct {
	byHash map[string]*domain.RefreshToken
	nextID int64
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{byHash: map[string]*domain.RefreshToken{}}
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, t *domain.RefreshToken) error {
	f.nextID++
	t.ID = f.nextID
	f.byHash[t.TokenHash] = t
	return nil
}
func (f *fakeRefreshTokenRepo) FindByHash(_ context.Context, h string) (*domain.RefreshToken, error) {
	return f.byHash[h], nil
}
func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id int64, replacedBy *int64) error {
	for _, t := range f.byHash {
		if t.ID == id {
			now := time.Now()
			t.RevokedAt = &now
			t.ReplacedByID = replacedBy
		}
	}
	return nil
}
func (f *fakeRefreshTokenRepo) RevokeAllForUser(context.Context, int64) error { return nil }

type fakeSessionRepo struct{}

func (fakeSessionRepo) Create(context.Context, *domain.Session) error { return nil }
func (fakeSessionRepo) ListActiveForUser(context.Context, int64) ([]domain.Session, error) {
	return nil, nil
}
func (fakeSessionRepo) Revoke(context.Context, int64) error { return nil }
func (fakeSessionRepo) Touch(context.Context, int64) error  { return nil }

func newTestAuthService() (*service.AuthService, *fakeUserRepo) {
	users := newFakeUserRepo()
	roles := newFakeRoleRepo()
	refresh := newFakeRefreshTokenRepo()
	issuer := jwtutil.NewIssuer("test-secret", 15*time.Minute, "test-issuer")
	return service.NewAuthService(users, roles, refresh, fakeSessionRepo{}, issuer, 30*24*time.Hour), users
}

func TestAuthService_Register_Success(t *testing.T) {
	svc, users := newTestAuthService()

	user, err := svc.Register(context.Background(), "Jane Doe", "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", user.Email)
	}
	if _, ok := users.byEmail["jane@example.com"]; !ok {
		t.Error("expected user to be persisted in repository")
	}
	if !hash.VerifyPassword(user.PasswordHash, "password123") {
		t.Error("expected password to be hashed and verifiable")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "Jane", "jane@example.com", "password123"); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err := svc.Register(ctx, "Impostor", "jane@example.com", "otherpassword")
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("expected an *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeConflict {
		t.Errorf("expected CodeConflict, got %s", appErr.Code)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "Jane", "jane@example.com", "correct-password"); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	_, err := svc.Login(ctx, "jane@example.com", "wrong-password", service.RequestMeta{})
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("expected an *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("expected CodeUnauthorized, got %s", appErr.Code)
	}
}

func TestAuthService_Login_Success_IssuesTokens(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "Jane", "jane@example.com", "correct-password"); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	tokens, err := svc.Login(ctx, "jane@example.com", "correct-password", service.RequestMeta{IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected both access and refresh tokens to be issued")
	}
	if tokens.User.Email != "jane@example.com" {
		t.Errorf("expected returned user to match, got %s", tokens.User.Email)
	}
}

func TestAuthService_Refresh_RotatesToken(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "Jane", "jane@example.com", "correct-password"); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	first, err := svc.Login(ctx, "jane@example.com", "correct-password", service.RequestMeta{})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	second, err := svc.Refresh(ctx, first.RefreshToken, service.RequestMeta{})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Error("expected refresh to rotate to a new refresh token, got the same one")
	}

	// The old (now-revoked) refresh token must no longer work.
	if _, err := svc.Refresh(ctx, first.RefreshToken, service.RequestMeta{}); err == nil {
		t.Error("expected the rotated-out refresh token to be rejected")
	}
}
