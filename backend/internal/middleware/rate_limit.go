package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/cache"
	"bookreader/backend/pkg/response"
)

// RateLimit enforces a fixed-window (per-minute) request cap per client IP,
// backed by Redis so the limit holds across multiple backend replicas rather
// than resetting per-process. A Redis outage fails open (never blocks
// legitimate traffic because caching is down) but logs nothing here by
// design — the caller already logs Redis connectivity issues at startup.
func RateLimit(c *cache.Cache, requestsPerMinute int) gin.HandlerFunc {
	return RateLimitScoped(c, "global", requestsPerMinute)
}

// RateLimitScoped is RateLimit with an explicit key scope, so a stricter
// per-endpoint limit (e.g. login, to blunt credential-stuffing/brute-force)
// doesn't share a counter — and therefore doesn't get exhausted by — the
// general per-IP traffic limit applied to every other route.
func RateLimitScoped(c *cache.Cache, scope string, requestsPerMinute int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if requestsPerMinute <= 0 {
			ctx.Next()
			return
		}

		window := time.Now().Unix() / 60
		key := fmt.Sprintf("ratelimit:%s:%s:%d", scope, ctx.ClientIP(), window)

		count, err := c.IncrWithExpiry(ctx.Request.Context(), key, time.Minute)
		if err != nil {
			ctx.Next() // fail open
			return
		}
		if count > int64(requestsPerMinute) {
			response.Fail(ctx, apperror.RateLimited("too many requests, please slow down"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
