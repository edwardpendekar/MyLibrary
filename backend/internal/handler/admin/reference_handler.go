package admin

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

type LanguageHandler struct{ svc *service.LanguageService }

func NewLanguageHandler(svc *service.LanguageService) *LanguageHandler {
	return &LanguageHandler{svc: svc}
}

// List godoc
// @Tags admin-languages
// @Success 200 {object} response.Envelope{data=[]dto.LanguageResponse}
// @Router /admin/languages [get]
func (h *LanguageHandler) List(c *gin.Context) {
	langs, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	out := make([]dto.LanguageResponse, 0, len(langs))
	for i := range langs {
		out = append(out, *dto.ToLanguageResponse(&langs[i]))
	}
	response.OK(c, out)
}

// Create godoc
// @Tags admin-languages
// @Accept json
// @Success 201 {object} response.Envelope{data=dto.LanguageResponse}
// @Router /admin/languages [post]
func (h *LanguageHandler) Create(c *gin.Context) {
	var req dto.CreateLanguageRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	lang := &domain.Language{Code: req.Code, Name: req.Name, NativeName: req.NativeName}
	if err := h.svc.Create(c.Request.Context(), lang); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, dto.ToLanguageResponse(lang))
}

// Update godoc
// @Tags admin-languages
// @Accept json
// @Param id path int true "language id"
// @Success 200 {object} response.Envelope{data=dto.LanguageResponse}
// @Router /admin/languages/{id} [put]
func (h *LanguageHandler) Update(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateLanguageRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	lang := &domain.Language{ID: id, Code: req.Code, Name: req.Name, NativeName: req.NativeName, IsActive: req.IsActive}
	if err := h.svc.Update(c.Request.Context(), lang); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToLanguageResponse(lang))
}

// Delete godoc
// @Tags admin-languages
// @Param id path int true "language id"
// @Success 204
// @Router /admin/languages/{id} [delete]
func (h *LanguageHandler) Delete(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List godoc
// @Tags admin-categories
// @Success 200 {object} response.Envelope{data=[]dto.CategoryResponse}
// @Router /admin/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	cats, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	out := make([]dto.CategoryResponse, 0, len(cats))
	for i := range cats {
		out = append(out, *dto.ToCategoryResponse(&cats[i]))
	}
	response.OK(c, out)
}

// Create godoc
// @Tags admin-categories
// @Accept json
// @Success 201 {object} response.Envelope{data=dto.CategoryResponse}
// @Router /admin/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	cat := &domain.Category{Slug: req.Slug, NameEN: req.NameEN, NameID: req.NameID, Description: req.Description, ParentID: req.ParentID}
	if err := h.svc.Create(c.Request.Context(), cat); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, dto.ToCategoryResponse(cat))
}

// Update godoc
// @Tags admin-categories
// @Accept json
// @Param id path int true "category id"
// @Success 200 {object} response.Envelope{data=dto.CategoryResponse}
// @Router /admin/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCategoryRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	cat := &domain.Category{ID: id, Slug: req.Slug, NameEN: req.NameEN, NameID: req.NameID, Description: req.Description, ParentID: req.ParentID}
	if err := h.svc.Update(c.Request.Context(), cat); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToCategoryResponse(cat))
}

// Delete godoc
// @Tags admin-categories
// @Param id path int true "category id"
// @Success 204
// @Router /admin/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}
