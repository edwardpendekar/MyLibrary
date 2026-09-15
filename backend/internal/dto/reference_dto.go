package dto

type CreateLanguageRequest struct {
	Code       string `json:"code" validate:"required,min=2,max=10"`
	Name       string `json:"name" validate:"required,max=100"`
	NativeName string `json:"native_name" validate:"required,max=100"`
}

type UpdateLanguageRequest struct {
	Code       string `json:"code" validate:"required,min=2,max=10"`
	Name       string `json:"name" validate:"required,max=100"`
	NativeName string `json:"native_name" validate:"required,max=100"`
	IsActive   bool   `json:"is_active"`
}

type CreateCategoryRequest struct {
	Slug        string `json:"slug" validate:"omitempty,max=150"`
	NameEN      string `json:"name_en" validate:"required,max=150"`
	NameID      string `json:"name_id" validate:"required,max=150"`
	Description string `json:"description"`
	ParentID    *int64 `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Slug        string `json:"slug" validate:"required,max=150"`
	NameEN      string `json:"name_en" validate:"required,max=150"`
	NameID      string `json:"name_id" validate:"required,max=150"`
	Description string `json:"description"`
	ParentID    *int64 `json:"parent_id"`
}
