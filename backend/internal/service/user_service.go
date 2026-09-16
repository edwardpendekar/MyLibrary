package service

import (
	"context"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/hash"
)

// UserService covers admin user management (list/update role/deactivate).
// Self-service registration lives in AuthService.
type UserService struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

func NewUserService(users domain.UserRepository, roles domain.RoleRepository) *UserService {
	return &UserService{users: users, roles: roles}
}

func (s *UserService) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.User], error) {
	return s.users.List(ctx, cursor, limit)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("failed to load user", err)
	}
	if user == nil {
		return nil, apperror.NotFound("user not found")
	}
	return user, nil
}

func (s *UserService) ChangeRole(ctx context.Context, userID int64, roleName string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	roles, err := s.roles.List(ctx)
	if err != nil {
		return apperror.Internal("failed to load roles", err)
	}
	for _, r := range roles {
		if r.Name == roleName {
			user.RoleID = r.ID
			return s.users.Update(ctx, user)
		}
	}
	return apperror.Validation("unknown role", map[string]string{"role": "must be one of admin, editor, user, guest"})
}

func (s *UserService) Deactivate(ctx context.Context, userID int64) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.IsActive = false
	return s.users.Update(ctx, user)
}

func (s *UserService) Activate(ctx context.Context, userID int64) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.IsActive = true
	return s.users.Update(ctx, user)
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, newPassword string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	hashed, err := hash.HashPassword(newPassword)
	if err != nil {
		return apperror.Internal("failed to hash password", err)
	}
	user.PasswordHash = hashed
	return s.users.Update(ctx, user)
}
