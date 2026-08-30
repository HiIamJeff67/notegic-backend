package validations

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterWarningValidation(validate *validator.Validate) {
	validate.RegisterValidation("iswarningcontent", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
}
