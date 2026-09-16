package middleware

import "github.com/gin-gonic/gin"

// SecureHeaders sets the standard defensive headers for a JSON API (the
// "Helmet" equivalent). The frontend (which serves the actual app HTML) sets
// its own, stricter CSP in next.config.ts; this one only needs to cover what
// the backend itself ever serves as HTML — the Swagger UI page — so it stays
// permissive enough for swagger-ui's inline bootstrap script/styles.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "0") // superseded by CSP; explicitly disabled per OWASP guidance
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		c.Next()
	}
}
