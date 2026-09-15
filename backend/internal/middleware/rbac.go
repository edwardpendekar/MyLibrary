package middleware

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/response"
)

// RequireRole must run after Auth. It 403s any request whose JWT role claim is
// not in the allowed set (e.g. admin-only routes: RequireRole(domain.RoleAdmin)).
func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := Role(c)
		if _, ok := allowedSet[role]; !ok {
			response.Fail(c, apperror.Forbidden("you do not have permission to perform this action"))
			c.Abort()
			return
		}
		c.Next()
	}
}
