// Package middleware holds cross-cutting Gin middleware: auth, RBAC, request
// logging, panic recovery, CORS, and (Phase 9) rate limiting / CSRF / secure headers.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/jwtutil"
	"bookreader/backend/pkg/response"
)

const (
	AccessTokenCookie = "access_token"
	ContextUserIDKey  = "auth.user_id"
	ContextUserEmail  = "auth.email"
	ContextUserRole   = "auth.role"
)

// Auth reads the access token from the httpOnly cookie set at login (the Next.js
// BFF route handlers forward it) or, as a fallback for non-browser API clients,
// from an Authorization: Bearer header. Missing/invalid token -> 401.
func Auth(issuer *jwtutil.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Fail(c, apperror.Unauthorized("authentication required"))
			c.Abort()
			return
		}

		claims, err := issuer.Parse(token)
		if err != nil {
			response.Fail(c, apperror.Unauthorized("session expired, please log in again"))
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

// OptionalAuth behaves like Auth but lets guests through with no context values
// set, for endpoints that personalize output (e.g. "is this book favorited?")
// without requiring a login.
func OptionalAuth(issuer *jwtutil.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		claims, err := issuer.Parse(token)
		if err != nil {
			c.Next()
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil && cookie != "" {
		return cookie
	}
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

// UserID returns the authenticated user's ID and whether one is present in context.
func UserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func Role(c *gin.Context) string {
	v, ok := c.Get(ContextUserRole)
	if !ok {
		return domain.RoleGuest
	}
	role, _ := v.(string)
	return role
}
