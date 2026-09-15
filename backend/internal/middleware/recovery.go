package middleware

import (
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/response"
)

// Recovery converts a panic anywhere down the handler chain into a clean 500
// JSON envelope instead of Gin's default plain-text crash dump, and logs the
// stack trace server-side so nothing is lost.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				dump, _ := httputil.DumpRequest(c.Request, false)
				log.Error().
					Interface("panic", rec).
					Str("request", string(dump)).
					Msg("panic recovered")
				response.Fail(c, apperror.Internal("internal server error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}
