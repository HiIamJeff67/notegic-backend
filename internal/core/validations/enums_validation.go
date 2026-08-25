package validation

import (
	"slices"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func RegisterEnumsValidation(validate *validator.Validate) {
	validate.RegisterValidation("isaccesscontrolpermission", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllAccessControlPermissionStrings, val)
	})
	validate.RegisterValidation("isbadgetype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllBadgeTypeStrings, val)
	})
	validate.RegisterValidation("isbillingintervalunit", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllBillingIntervalUnitStrings, val)
	})
	validate.RegisterValidation("isbillingplanname", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllBillingPlanNameStrings, val)
	})
	validate.RegisterValidation("isbillingplanstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllBillingPlanStatusStrings, val)
	})
	validate.RegisterValidation("isblocktype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllBlockTypeStrings, val)
	})
	validate.RegisterValidation("iscountrycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllCountryCodeStrings, val)
	})
	validate.RegisterValidation("iscountry", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllCountryStrings, val)
	})
	validate.RegisterValidation("isitemtype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllItemTypeStrings, val)
	})
	validate.RegisterValidation("islanguage", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllLanguageStrings, val)
	})
	validate.RegisterValidation("ismaterialcontenttype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllMaterialContentTypeStrings, val)
	})
	validate.RegisterValidation("isroutineperiod", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllRoutinePeriodStrings, val)
	})
	validate.RegisterValidation("isroutinestatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllRoutineStatusStrings, val)
	})
	validate.RegisterValidation("isroutinetaskpurpose", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllRoutineTaskPurposeStrings, val)
	})
	validate.RegisterValidation("isroutinetaskstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllRoutineTaskStatusStrings, val)
	})
	validate.RegisterValidation("issupportedicon", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllSupportedIconStrings, val)
	})
	validate.RegisterValidation("issupportedcurrencycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllSupportedCurrencyCodeStrings, val)
	})
	validate.RegisterValidation("isgender", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllUserGenderStrings, val)
	})
	validate.RegisterValidation("isplan", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllUserPlanStrings, val)
	})
	validate.RegisterValidation("isrole", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllUserRoleStrings, val)
	})
	validate.RegisterValidation("isstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllUserStatusStrings, val)
	})
	validate.RegisterValidation("isuserstobillingplansstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(enums.AllUsersToBillingPlansStatusStrings, val)
	})
}
