package validation

import (
	"github.com/go-playground/validator/v10"

	sharedvalidation "github.com/HiIamJeff67/notegic-backend/shared/validations"

	cblocknotevalidations "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote/validations"
)

func New() *validator.Validate {
	validator := validator.New()
	cblocknotevalidations.RegisterShelfBlockValidation(validator)
	sharedvalidation.RegisterStringsValidation(validator)
	sharedvalidation.RegisterTimesValidation(validator)
	RegisterEnumsValidation(validator)
	return validator
}
