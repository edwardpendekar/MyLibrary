package middleware

import "github.com/gin-gonic/gin"

// SecureHeaders sets the standard defensive headers for a JSON API (the
// "Helmet" equivalent). CSP is left to the frontend, which serves the actual HTML.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "0") // superseded by CSP; explicitly disabled per OWASP guidance
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}
