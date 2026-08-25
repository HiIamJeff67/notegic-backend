package validation

import (
	"github.com/go-playground/validator/v10"

	cblocknotevalidations "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote/validations"

	svalidations "github.com/HiIamJeff67/notegic-backend/shared/validations"
)

func New() *validator.Validate {
	validator := validator.New()
	cblocknotevalidations.RegisterShelfBlockValidation(validator)
	svalidations.RegisterStringsValidation(validator)
	svalidations.RegisterTimesValidation(validator)
	RegisterEnumsValidation(validator)
	return validator
}
