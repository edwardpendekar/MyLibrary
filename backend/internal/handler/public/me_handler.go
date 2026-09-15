package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/response"
)

type MeHandler struct{ users *service.UserService }

func NewMeHandler(users *service.UserService) *MeHandler { return &MeHandler{users: users} }

// Me godoc
// @Summary      Current authenticated user's profile
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.Envelope{data=dto.UserResponse}
// @Router       /me [get]
func (h *MeHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, apperror.Unauthorized("authentication required"))
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToUserResponse(user))
}
