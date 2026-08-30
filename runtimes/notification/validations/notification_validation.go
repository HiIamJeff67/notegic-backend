package validations

import "github.com/go-playground/validator/v10"

func RegisterNotificationValidation(validate *validator.Validate) {
	validate.RegisterValidation("isnotificationtype", func(fl validator.FieldLevel) bool {
		switch fl.Field().String() {
		case "news", "warning", "important":
			return true
		default:
			return false
		}
	})
	validate.RegisterValidation("isnotificationpriority", func(fl validator.FieldLevel) bool {
		switch fl.Field().String() {
		case "low", "normal", "high", "critical":
			return true
		default:
			return false
		}
	})
}
