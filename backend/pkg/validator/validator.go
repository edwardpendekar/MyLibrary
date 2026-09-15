// Package validator wraps go-playground/validator so handlers get back a
// field->message map instead of raw validator.FieldError values.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var instance = validator.New(validator.WithRequiredStructEnabled())

// Struct validates s and, on failure, returns a map of field name (JSON-ish, lowercased)
// to a human-readable message suitable for apperror.Validation.
func Struct(s interface{}) map[string]string {
	err := instance.Struct(s)
	if err == nil {
		return nil
	}

	fields := make(map[string]string)
	var verrs validator.ValidationErrors
	if !isValidationErrors(err, &verrs) {
		fields["_"] = err.Error()
		return fields
	}

	for _, fe := range verrs {
		fields[strings.ToLower(fe.Field())] = message(fe)
	}
	return fields
}

func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	verrs, ok := err.(validator.ValidationErrors)
	if ok {
		*target = verrs
	}
	return ok
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
