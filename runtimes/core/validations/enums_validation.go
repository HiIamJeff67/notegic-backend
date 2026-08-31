package validation

import (
	"slices"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func RegisterEnumsValidation(validate *validator.Validate) {
	validate.RegisterValidation("isaccesscontrolpermission", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllAccessControlPermissionStrings, val)
	})
	validate.RegisterValidation("isbadgetype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllBadgeTypeStrings, val)
	})
	validate.RegisterValidation("isbillingintervalunit", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllBillingIntervalUnitStrings, val)
	})
	validate.RegisterValidation("isbillingplanname", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllBillingPlanNameStrings, val)
	})
	validate.RegisterValidation("isbillingplanstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllBillingPlanStatusStrings, val)
	})
	validate.RegisterValidation("isblocktype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllBlockTypeStrings, val)
	})
	validate.RegisterValidation("iscountrycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllCountryCodeStrings, val)
	})
	validate.RegisterValidation("iscountry", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllCountryStrings, val)
	})
	validate.RegisterValidation("isitemtype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllItemTypeStrings, val)
	})
	validate.RegisterValidation("islanguage", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllLanguageStrings, val)
	})
	validate.RegisterValidation("ismaterialcontenttype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllMaterialContentTypeStrings, val)
	})
	validate.RegisterValidation("isroutineperiod", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllRoutinePeriodStrings, val)
	})
	validate.RegisterValidation("isroutinestatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllRoutineStatusStrings, val)
	})
	validate.RegisterValidation("isroutinetaskpurpose", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllRoutineTaskPurposeStrings, val)
	})
	validate.RegisterValidation("issupportedicon", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllSupportedIconStrings, val)
	})
	validate.RegisterValidation("issupportedcurrencycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllSupportedCurrencyCodeStrings, val)
	})
	validate.RegisterValidation("isgender", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllUserGenderStrings, val)
	})
	validate.RegisterValidation("isplan", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllUserPlanStrings, val)
	})
	validate.RegisterValidation("isrole", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllUserRoleStrings, val)
	})
	validate.RegisterValidation("isstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllUserStatusStrings, val)
	})
	validate.RegisterValidation("isuserstobillingplansstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(cenums.AllUsersToBillingPlansStatusStrings, val)
	})
}
