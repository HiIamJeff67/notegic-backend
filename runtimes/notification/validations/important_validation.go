package validations

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterImportantValidation(validate *validator.Validate) {
	validate.RegisterValidation("isimportantcontent", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
}
