package middleware

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/service"
)

// Audit wraps a write endpoint and, only if it succeeded (2xx), writes an
// audit_logs row. entityType is fixed per route; the entity ID is taken from the
// route's :id param when present (e.g. PUT /admin/books/:id).
func Audit(auditSvc *service.AuditService, action, entityType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		var userID *int64
		if id, ok := UserID(c); ok {
			userID = &id
		}
		var entityID *string
		if id := c.Param("id"); id != "" {
			entityID = &id
		}

		_ = auditSvc.Record(c.Request.Context(), service.RecordInput{
			UserID: userID, Action: action, EntityType: entityType, EntityID: entityID,
			IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		})
	}
}
