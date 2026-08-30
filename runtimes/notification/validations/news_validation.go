package validations

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterNewsValidation(validate *validator.Validate) {
	validate.RegisterValidation("isnewscontent", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
}
