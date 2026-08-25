package validation

import (
	"slices"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

var (
	allAccessControlPermissionStrings = []string{
		string(enums.AccessControlPermission_Read),
		string(enums.AccessControlPermission_Write),
		string(enums.AccessControlPermission_Admin),
		string(enums.AccessControlPermission_Owner),
	}
	allBadgeTypeStrings = []string{
		string(enums.BadgeType_Diamond),
		string(enums.BadgeType_Golden),
		string(enums.BadgeType_Silver),
		string(enums.BadgeType_Bronze),
		string(enums.BadgeType_Steel),
	}
	allBillingIntervalUnitStrings = []string{
		string(enums.BillingIntervalUnit_Day),
		string(enums.BillingIntervalUnit_Week),
		string(enums.BillingIntervalUnit_Month),
		string(enums.BillingIntervalUnit_Year),
	}
	allBillingPlanNameStrings = []string{
		string(enums.BillingPlanName_NotegicMonthlyFreePlan),
		string(enums.BillingPlanName_NotegicMonthlyProPlan),
		string(enums.BillingPlanName_NotegicYearlyProPlan),
		string(enums.BillingPlanName_NotegicMonthlyPremiumPlan),
		string(enums.BillingPlanName_NotegicYearlyPremiumPlan),
		string(enums.BillingPlanName_NotegicMonthlyUltimatePlan),
		string(enums.BillingPlanName_NotegicYearlyUltimatePlan),
		string(enums.BillingPlanName_NotegicMonthlyEnterprisePlan),
		string(enums.BillingPlanName_NotegicYearlyEnterprisePlan),
	}
	allBillingPlanStatusStrings = []string{
		string(enums.BillingPlanStatus_Created),
		string(enums.BillingPlanStatus_Active),
		string(enums.BillingPlanStatus_Inactive),
	}
	allBlockTypeStrings = []string{
		string(enums.BlockType_Paragraph),
		string(enums.BlockType_Heading),
		string(enums.BlockType_Quote),
		string(enums.BlockType_BulletListItem),
		string(enums.BlockType_NumberedListItem),
		string(enums.BlockType_CheckListItem),
		string(enums.BlockType_ToggleListItem),
		string(enums.BlockType_Image),
		string(enums.BlockType_Video),
		string(enums.BlockType_Audio),
		string(enums.BlockType_File),
		string(enums.BlockType_Table),
		string(enums.BlockType_CodeBlock),
	}
	allCountryCodeStrings = []string{
		string(enums.CountryCode_Taiwan),
		string(enums.CountryCode_Japan),
		string(enums.CountryCode_Malaysia),
		string(enums.CountryCode_Singapore),
		string(enums.CountryCode_China),
		string(enums.CountryCode_NANP),
		string(enums.CountryCode_UnitedKingdom),
		string(enums.CountryCode_Australia),
	}
	allCountryStrings = []string{
		string(enums.Country_Taiwan),
		string(enums.Country_Japan),
		string(enums.Country_Malaysia),
		string(enums.Country_Singapore),
		string(enums.Country_China),
		string(enums.Country_UnitedStatusOfAmerica),
		string(enums.Country_UnitedKingdom),
		string(enums.Country_Australia),
		string(enums.Country_Canada),
	}
	allItemTypeStrings = []string{
		string(enums.ItemType_BlockPack),
		string(enums.ItemType_Material),
	}
	allLanguageStrings = []string{
		string(enums.Language_English),
		string(enums.Language_TraditionalChinese),
		string(enums.Language_SimpleChinese),
		string(enums.Language_Japanese),
		string(enums.Language_Korean),
	}
	allMaterialContentTypeStrings = []string{
		string(enums.MaterialContentType_None),
		string(enums.MaterialContentType_JSON),
		string(enums.MaterialContentType_PDF),
		string(enums.MaterialContentType_PlainText),
		string(enums.MaterialContentType_HTML),
		string(enums.MaterialContentType_Markdown),
		string(enums.MaterialContentType_PNG),
		string(enums.MaterialContentType_JPG),
		string(enums.MaterialContentType_JPEG),
		string(enums.MaterialContentType_GIF),
		string(enums.MaterialContentType_SVG),
		string(enums.MaterialContentType_WebP),
		string(enums.MaterialContentType_MP4),
		string(enums.MaterialContentType_WebM),
		string(enums.MaterialContentType_Mpeg),
	}
	allRoutinePeriodStrings = []string{
		string(enums.RoutinePeriod_Daily),
		string(enums.RoutinePeriod_Weekly),
		string(enums.RoutinePeriod_Monthly),
	}
	allRoutineStatusStrings = []string{
		string(enums.RoutineStatus_Scheduled),
		string(enums.RoutineStatus_InProgress),
		string(enums.RoutineStatus_Completed),
		string(enums.RoutineStatus_OverDue),
	}
	allRoutineTaskPurposeStrings = []string{
		string(enums.RoutineTaskPurpose_CreateRootShelf),
		string(enums.RoutineTaskPurpose_UpdateRootShelf),
		string(enums.RoutineTaskPurpose_ResetRootShelf),
		string(enums.RoutineTaskPurpose_CreateSubShelf),
		string(enums.RoutineTaskPurpose_UpdateSubShelf),
		string(enums.RoutineTaskPurpose_ResetSubShelf),
		string(enums.RoutineTaskPurpose_CreateBlockPack),
		string(enums.RoutineTaskPurpose_UpdateBlockPack),
		string(enums.RoutineTaskPurpose_ResetBlockPack),
		string(enums.RoutineTaskPurpose_AppendBlock),
		string(enums.RoutineTaskPurpose_UpdateBlock),
		string(enums.RoutineTaskPurpose_ResetBlock),
		string(enums.RoutineTaskPurpose_CreateRoutine),
		string(enums.RoutineTaskPurpose_UpdateRoutine),
	}
	allRoutineTaskStatusStrings = []string{
		string(enums.RoutineTaskStatus_Idle),
		string(enums.RoutineTaskStatus_Waiting),
		string(enums.RoutineTaskStatus_Running),
		string(enums.RoutineTaskStatus_Pause),
	}
	allSupportedIconStrings = []string{
		string(enums.SupportedIcon_GrinningFace),
		string(enums.SupportedIcon_SmilingFaceWithSmilingEyes),
		string(enums.SupportedIcon_RedHeart),
		string(enums.SupportedIcon_Fire),
		string(enums.SupportedIcon_Star),
		string(enums.SupportedIcon_Books),
		string(enums.SupportedIcon_Notebook),
		string(enums.SupportedIcon_PencilPaper),
		string(enums.SupportedIcon_Lightbulb),
		string(enums.SupportedIcon_Rocket),
		string(enums.SupportedIcon_CheckMark),
		string(enums.SupportedIcon_Pin),
		string(enums.SupportedIcon_FolderOpen),
		string(enums.SupportedIcon_Calendar),
		string(enums.SupportedIcon_Clock),
	}
	allSupportedCurrencyCodeStrings = []string{
		string(enums.SupportedCurrencyCode_USD),
		string(enums.SupportedCurrencyCode_EUR),
		string(enums.SupportedCurrencyCode_JPY),
		string(enums.SupportedCurrencyCode_TWD),
		string(enums.SupportedCurrencyCode_KRW),
		string(enums.SupportedCurrencyCode_CNY),
	}
	allUserGenderStrings = []string{
		string(enums.UserGender_Male),
		string(enums.UserGender_Female),
		string(enums.UserGender_PreferNotToSay),
	}
	allUserPlanStrings = []string{
		string(enums.UserPlan_Enterprise),
		string(enums.UserPlan_Ultimate),
		string(enums.UserPlan_Premium),
		string(enums.UserPlan_Pro),
		string(enums.UserPlan_Free),
	}
	allUserRoleStrings = []string{
		string(enums.UserRole_Admin),
		string(enums.UserRole_Normal),
		string(enums.UserRole_Guest),
	}
	allUserStatusStrings = []string{
		string(enums.UserStatus_Online),
		string(enums.UserStatus_AFK),
		string(enums.UserStatus_DoNotDisturb),
		string(enums.UserStatus_Offline),
	}
	allUsersToBillingPlansStatusStrings = []string{
		string(enums.UsersToBillingPlansStatus_ApprovalPending),
		string(enums.UsersToBillingPlansStatus_Approved),
		string(enums.UsersToBillingPlansStatus_Active),
		string(enums.UsersToBillingPlansStatus_Suspended),
		string(enums.UsersToBillingPlansStatus_Cancelled),
		string(enums.UsersToBillingPlansStatus_Expired),
	}
)

func RegisterEnumsValidation(validate *validator.Validate) {
	validate.RegisterValidation("isaccesscontrolpermission", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allAccessControlPermissionStrings, val)
	})
	validate.RegisterValidation("isbadgetype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allBadgeTypeStrings, val)
	})
	validate.RegisterValidation("isbillingintervalunit", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allBillingIntervalUnitStrings, val)
	})
	validate.RegisterValidation("isbillingplanname", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allBillingPlanNameStrings, val)
	})
	validate.RegisterValidation("isbillingplanstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allBillingPlanStatusStrings, val)
	})
	validate.RegisterValidation("isblocktype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allBlockTypeStrings, val)
	})
	validate.RegisterValidation("iscountrycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allCountryCodeStrings, val)
	})
	validate.RegisterValidation("iscountry", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allCountryStrings, val)
	})
	validate.RegisterValidation("isitemtype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allItemTypeStrings, val)
	})
	validate.RegisterValidation("islanguage", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allLanguageStrings, val)
	})
	validate.RegisterValidation("ismaterialcontenttype", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allMaterialContentTypeStrings, val)
	})
	validate.RegisterValidation("isroutineperiod", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allRoutinePeriodStrings, val)
	})
	validate.RegisterValidation("isroutinestatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allRoutineStatusStrings, val)
	})
	validate.RegisterValidation("isroutinetaskpurpose", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allRoutineTaskPurposeStrings, val)
	})
	validate.RegisterValidation("isroutinetaskstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allRoutineTaskStatusStrings, val)
	})
	validate.RegisterValidation("issupportedicon", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allSupportedIconStrings, val)
	})
	validate.RegisterValidation("issupportedcurrencycode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allSupportedCurrencyCodeStrings, val)
	})
	validate.RegisterValidation("isgender", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allUserGenderStrings, val)
	})
	validate.RegisterValidation("isplan", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allUserPlanStrings, val)
	})
	validate.RegisterValidation("isrole", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allUserRoleStrings, val)
	})
	validate.RegisterValidation("isstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allUserStatusStrings, val)
	})
	validate.RegisterValidation("isuserstobillingplansstatus", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return slices.Contains(allUsersToBillingPlansStatusStrings, val)
	})
}
