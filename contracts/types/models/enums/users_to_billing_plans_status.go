package enums

type UsersToBillingPlansStatus string

const (
	UsersToBillingPlansStatus_ApprovalPending UsersToBillingPlansStatus = "APPROVAL_PENDING"
	UsersToBillingPlansStatus_Approved        UsersToBillingPlansStatus = "APPROVED"
	UsersToBillingPlansStatus_Active          UsersToBillingPlansStatus = "ACTIVE"
	UsersToBillingPlansStatus_Suspended       UsersToBillingPlansStatus = "SUSPENDED"
	UsersToBillingPlansStatus_Cancelled       UsersToBillingPlansStatus = "CANCELLED"
	UsersToBillingPlansStatus_Expired         UsersToBillingPlansStatus = "EXPIRED"
)
