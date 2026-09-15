package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const ContextRequestIDKey = "request_id"

// RequestID stamps every request with a correlation ID, echoed back in the
// X-Request-ID response header so client-side error reports can be matched to
// server-side logs.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ContextRequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestLogger emits one structured log line per request via zerolog, tagged
// with the request ID so it can be grepped alongside the client-reported ID.
func RequestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(ContextRequestIDKey)
		log.Info().
			Str("request_id", asString(requestID)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("http_request")
	}
}

func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}
