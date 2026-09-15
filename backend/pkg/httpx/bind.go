// Package httpx holds small Gin request-parsing helpers shared by every handler
// package, so binding/validation error handling is written exactly once.
package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/response"
	"bookreader/backend/pkg/validator"
)

// BindJSON decodes the request body into dst and runs struct validation tags.
// On failure it writes the error response itself and returns false so the
// caller can simply `if !httpx.BindJSON(c, &req) { return }`.
func BindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Fail(c, apperror.Validation("invalid request body", map[string]string{"_": err.Error()}))
		return false
	}
	if fields := validator.Struct(dst); fields != nil {
		response.Fail(c, apperror.Validation("validation failed", fields))
		return false
	}
	return true
}

// ParamInt64 parses a required :id-style path parameter.
func ParamInt64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		response.Fail(c, apperror.Validation("invalid "+name, nil))
		return 0, false
	}
	return id, true
}

func QueryInt(c *gin.Context, name string, fallback int) int {
	raw := c.Query(name)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func QueryInt64Ptr(c *gin.Context, name string) *int64 {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}
