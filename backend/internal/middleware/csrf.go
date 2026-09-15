package middleware

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"

	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/csrf"
	"bookreader/backend/pkg/response"
)

var safeMethods = map[string]bool{"GET": true, "HEAD": true, "OPTIONS": true}

// CSRF rejects state-changing requests from an authenticated session (one
// carrying the access-token cookie) whose X-CSRF-Token header doesn't match
// the csrf_token cookie. Requests with no session cookie at all (a fresh
// login/register call) are exempt — there is no session to forge yet.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if safeMethods[c.Request.Method] {
			c.Next()
			return
		}

		sessionCookie, err := c.Cookie(AccessTokenCookie)
		if err != nil || sessionCookie == "" {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(csrf.CookieName)
		if err != nil || cookieToken == "" {
			response.Fail(c, apperror.Forbidden("missing CSRF token"))
			c.Abort()
			return
		}

		headerToken := c.GetHeader(csrf.HeaderName)
		if headerToken == "" || subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
			response.Fail(c, apperror.Forbidden("invalid CSRF token"))
			c.Abort()
			return
		}
		c.Next()
	}
}
