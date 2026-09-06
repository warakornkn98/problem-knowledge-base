// Package validatorx wraps go-playground/validator with a single shared
// instance and turns validation failures into readable messages.
package validatorx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/team/pkb/internal/shared/httpx"
)

var validate = validator.New()

// Struct validates s and, on failure, returns an *httpx.AppError (422) whose
// message lists the offending fields.
func Struct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return httpx.NewInternal(err)
	}

	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return httpx.NewInternal(err)
	}

	msgs := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		msgs = append(msgs, describe(fe))
	}
	return httpx.NewUnprocessable(strings.Join(msgs, "; "))
}

func describe(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, fe.Param())
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "gte":
		return fmt.Sprintf("%s must be >= %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, fe.Tag())
	}
}
