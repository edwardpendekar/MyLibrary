package admin

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/response"
)

type StatsHandler struct{ svc *service.StatsService }

func NewStatsHandler(svc *service.StatsService) *StatsHandler { return &StatsHandler{svc: svc} }

// Dashboard godoc
// @Summary      Admin dashboard summary counters
// @Tags         admin-stats
// @Produce      json
// @Success      200 {object} response.Envelope{data=dto.DashboardStatsResponse}
// @Router       /admin/stats [get]
func (h *StatsHandler) Dashboard(c *gin.Context) {
	stats, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToDashboardStatsResponse(stats))
}
