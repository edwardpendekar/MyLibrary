// Package response provides a single JSON envelope shape for every API response,
// success or error, so frontend clients can parse them uniformly.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bookreader/backend/pkg/apperror"
)

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func OKWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

// Fail maps an error to an HTTP status + envelope. Unknown error types are treated
// as internal errors so no accidental internal detail leaks to the client.
func Fail(c *gin.Context, err error) {
	appErr, ok := apperror.As(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, Envelope{
			Success: false,
			Error:   &ErrorBody{Code: string(apperror.CodeInternal), Message: "internal server error"},
		})
		return
	}

	status := statusFor(appErr.Code)
	body := &ErrorBody{Code: string(appErr.Code), Message: appErr.Message, Fields: appErr.Fields}
	c.JSON(status, Envelope{Success: false, Error: body})
}

func statusFor(code apperror.Code) int {
	switch code {
	case apperror.CodeValidation:
		return http.StatusUnprocessableEntity
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
