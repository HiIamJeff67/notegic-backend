package validation

import (
	"slices"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

var (
	allAccessControlPermissionStrings = []string{
		string(cenums.AccessControlPermission_Read),
		string(cenums.AccessControlPermission_Write),
		string(cenums.AccessControlPermission_Admin),
		string(cenums.AccessControlPermission_Owner),
	}
	allBadgeTypeStrings = []string{
		string(cenums.BadgeType_Diamond),
		string(cenums.BadgeType_Golden),
		string(cenums.BadgeType_Silver),
		string(cenums.BadgeType_Bronze),
		string(cenums.BadgeType_Steel),
	}
	allBillingIntervalUnitStrings = []string{
		string(cenums.BillingIntervalUnit_Day),
		string(cenums.BillingIntervalUnit_Week),
		string(cenums.BillingIntervalUnit_Month),
		string(cenums.BillingIntervalUnit_Year),
	}
	allBillingPlanNameStrings = []string{
		string(cenums.BillingPlanName_NotegicMonthlyFreePlan),
		string(cenums.BillingPlanName_NotegicMonthlyProPlan),
		string(cenums.BillingPlanName_NotegicYearlyProPlan),
		string(cenums.BillingPlanName_NotegicMonthlyPremiumPlan),
		string(cenums.BillingPlanName_NotegicYearlyPremiumPlan),
		string(cenums.BillingPlanName_NotegicMonthlyUltimatePlan),
		string(cenums.BillingPlanName_NotegicYearlyUltimatePlan),
		string(cenums.BillingPlanName_NotegicMonthlyEnterprisePlan),
		string(cenums.BillingPlanName_NotegicYearlyEnterprisePlan),
	}
	allBillingPlanStatusStrings = []string{
		string(cenums.BillingPlanStatus_Created),
		string(cenums.BillingPlanStatus_Active),
		string(cenums.BillingPlanStatus_Inactive),
	}
	allBlockTypeStrings = []string{
		string(cenums.BlockType_Paragraph),
		string(cenums.BlockType_Heading),
		string(cenums.BlockType_Quote),
		string(cenums.BlockType_BulletListItem),
		string(cenums.BlockType_NumberedListItem),
		string(cenums.BlockType_CheckListItem),
		string(cenums.BlockType_ToggleListItem),
		string(cenums.BlockType_Image),
		string(cenums.BlockType_Video),
		string(cenums.BlockType_Audio),
		string(cenums.BlockType_File),
		string(cenums.BlockType_Table),
		string(cenums.BlockType_CodeBlock),
	}
	allCountryCodeStrings = []string{
		string(cenums.CountryCode_Taiwan),
		string(cenums.CountryCode_Japan),
		string(cenums.CountryCode_Malaysia),
		string(cenums.CountryCode_Singapore),
		string(cenums.CountryCode_China),
		string(cenums.CountryCode_NANP),
		string(cenums.CountryCode_UnitedKingdom),
		string(cenums.CountryCode_Australia),
	}
	allCountryStrings = []string{
		string(cenums.Country_Taiwan),
		string(cenums.Country_Japan),
		string(cenums.Country_Malaysia),
		string(cenums.Country_Singapore),
		string(cenums.Country_China),
		string(cenums.Country_UnitedStatusOfAmerica),
		string(cenums.Country_UnitedKingdom),
		string(cenums.Country_Australia),
		string(cenums.Country_Canada),
	}
	allItemTypeStrings = []string{
		string(cenums.ItemType_BlockPack),
		string(cenums.ItemType_Material),
	}
	allLanguageStrings = []string{
		string(cenums.Language_English),
		string(cenums.Language_TraditionalChinese),
		string(cenums.Language_SimpleChinese),
		string(cenums.Language_Japanese),
		string(cenums.Language_Korean),
	}
	allMaterialContentTypeStrings = []string{
		string(cenums.MaterialContentType_None),
		string(cenums.MaterialContentType_JSON),
		string(cenums.MaterialContentType_PDF),
		string(cenums.MaterialContentType_PlainText),
		string(cenums.MaterialContentType_HTML),
		string(cenums.MaterialContentType_Markdown),
		string(cenums.MaterialContentType_PNG),
		string(cenums.MaterialContentType_JPG),
		string(cenums.MaterialContentType_JPEG),
		string(cenums.MaterialContentType_GIF),
		string(cenums.MaterialContentType_SVG),
		string(cenums.MaterialContentType_WebP),
		string(cenums.MaterialContentType_MP4),
		string(cenums.MaterialContentType_WebM),
		string(cenums.MaterialContentType_Mpeg),
	}
	allRoutinePeriodStrings = []string{
		string(cenums.RoutinePeriod_Daily),
		string(cenums.RoutinePeriod_Weekly),
		string(cenums.RoutinePeriod_Monthly),
	}
	allRoutineStatusStrings = []string{
		string(cenums.RoutineStatus_Scheduled),
		string(cenums.RoutineStatus_InProgress),
		string(cenums.RoutineStatus_Completed),
		string(cenums.RoutineStatus_OverDue),
	}
	allRoutineTaskPurposeStrings = []string{
		string(cenums.RoutineTaskPurpose_CreateRootShelf),
		string(cenums.RoutineTaskPurpose_UpdateRootShelf),
		string(cenums.RoutineTaskPurpose_ResetRootShelf),
		string(cenums.RoutineTaskPurpose_CreateSubShelf),
		string(cenums.RoutineTaskPurpose_UpdateSubShelf),
		string(cenums.RoutineTaskPurpose_ResetSubShelf),
		string(cenums.RoutineTaskPurpose_CreateBlockPack),
		string(cenums.RoutineTaskPurpose_UpdateBlockPack),
		string(cenums.RoutineTaskPurpose_ResetBlockPack),
		string(cenums.RoutineTaskPurpose_AppendBlock),
		string(cenums.RoutineTaskPurpose_UpdateBlock),
		string(cenums.RoutineTaskPurpose_ResetBlock),
		string(cenums.RoutineTaskPurpose_CreateRoutine),
		string(cenums.RoutineTaskPurpose_UpdateRoutine),
	}
	allRoutineTaskStatusStrings = []string{
		string(cenums.RoutineTaskStatus_Idle),
		string(cenums.RoutineTaskStatus_Waiting),
		string(cenums.RoutineTaskStatus_Running),
		string(cenums.RoutineTaskStatus_Pause),
	}
	allSupportedIconStrings = []string{
		string(cenums.SupportedIcon_GrinningFace),
		string(cenums.SupportedIcon_SmilingFaceWithSmilingEyes),
		string(cenums.SupportedIcon_RedHeart),
		string(cenums.SupportedIcon_Fire),
		string(cenums.SupportedIcon_Star),
		string(cenums.SupportedIcon_Books),
		string(cenums.SupportedIcon_Notebook),
		string(cenums.SupportedIcon_PencilPaper),
		string(cenums.SupportedIcon_Lightbulb),
		string(cenums.SupportedIcon_Rocket),
		string(cenums.SupportedIcon_CheckMark),
		string(cenums.SupportedIcon_Pin),
		string(cenums.SupportedIcon_FolderOpen),
		string(cenums.SupportedIcon_Calendar),
		string(cenums.SupportedIcon_Clock),
	}
	allSupportedCurrencyCodeStrings = []string{
		string(cenums.SupportedCurrencyCode_USD),
		string(cenums.SupportedCurrencyCode_EUR),
		string(cenums.SupportedCurrencyCode_JPY),
		string(cenums.SupportedCurrencyCode_TWD),
		string(cenums.SupportedCurrencyCode_KRW),
		string(cenums.SupportedCurrencyCode_CNY),
	}
	allUserGenderStrings = []string{
		string(cenums.UserGender_Male),
		string(cenums.UserGender_Female),
		string(cenums.UserGender_PreferNotToSay),
	}
	allUserPlanStrings = []string{
		string(cenums.UserPlan_Enterprise),
		string(cenums.UserPlan_Ultimate),
		string(cenums.UserPlan_Premium),
		string(cenums.UserPlan_Pro),
		string(cenums.UserPlan_Free),
	}
	allUserRoleStrings = []string{
		string(cenums.UserRole_Admin),
		string(cenums.UserRole_Normal),
		string(cenums.UserRole_Guest),
	}
	allUserStatusStrings = []string{
		string(cenums.UserStatus_Online),
		string(cenums.UserStatus_AFK),
		string(cenums.UserStatus_DoNotDisturb),
		string(cenums.UserStatus_Offline),
	}
	allUsersToBillingPlansStatusStrings = []string{
		string(cenums.UsersToBillingPlansStatus_ApprovalPending),
		string(cenums.UsersToBillingPlansStatus_Approved),
		string(cenums.UsersToBillingPlansStatus_Active),
		string(cenums.UsersToBillingPlansStatus_Suspended),
		string(cenums.UsersToBillingPlansStatus_Cancelled),
		string(cenums.UsersToBillingPlansStatus_Expired),
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
