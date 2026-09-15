package admin

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// List godoc
// @Tags admin-users
// @Success 200 {object} response.Envelope{data=[]dto.UserResponse}
// @Router /admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
	result, err := h.svc.List(c.Request.Context(), c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := make([]dto.UserResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, dto.ToUserResponse(&result.Items[i]))
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

// ChangeRole godoc
// @Tags admin-users
// @Accept json
// @Param id path int true "user id"
// @Success 204
// @Router /admin/users/{id}/role [put]
func (h *UserHandler) ChangeRole(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.ChangeRoleRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	if err := h.svc.ChangeRole(c.Request.Context(), id, req.Role); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// Deactivate godoc
// @Tags admin-users
// @Param id path int true "user id"
// @Success 204
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Deactivate(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Deactivate(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}
