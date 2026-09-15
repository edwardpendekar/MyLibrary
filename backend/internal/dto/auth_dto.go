// Package dto holds request/response shapes for the HTTP boundary. Keeping
// these separate from domain entities means changing an API contract never
// forces a change to business logic, and vice versa.
package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=150"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	User                 UserResponse `json:"user"`
	AccessTokenExpiresAt time.Time    `json:"access_token_expires_at"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type ChangeRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin editor user guest"`
}
