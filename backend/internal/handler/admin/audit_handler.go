package admin

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

type AuditHandler struct{ svc *service.AuditService }

func NewAuditHandler(svc *service.AuditService) *AuditHandler { return &AuditHandler{svc: svc} }

// List godoc
// @Summary      List audit log entries
// @Tags         admin-audit
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /admin/audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	result, err := h.svc.List(c.Request.Context(), c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := make([]dto.AuditLogResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, dto.ToAuditLogResponse(&result.Items[i]))
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}
