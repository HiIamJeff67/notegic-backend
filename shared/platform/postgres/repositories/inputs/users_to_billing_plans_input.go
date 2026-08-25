package inputs

import (
	"time"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateUsersToBillingPlansInput struct {
	BillingPlanId   string                           `json:"billingPlanId" gorm:"column:billing_plan_id;"`
	Status          cenums.UsersToBillingPlansStatus `json:"status" gorm:"column:status;"`
	StartDate       time.Time                        `json:"startTime" gorm:"column:start_date;"`
	EndDate         *time.Time                       `json:"endDate" gorm:"column:end_date;"`
	NextBillingDate time.Time                        `json:"nextBillingDate" gorm:"column:next_billing_date;"`
	FailureCount    int32                            `json:"failureCount" gorm:"column:failure_count;"`
	CreatedAt       time.Time                        `json:"createdAt" gorm:"column:created_at;"`
	UpdatedAt       time.Time                        `json:"updatedAt" gorm:"column:updated_at;"`
}

type UpdateUsersToBillingPlansInput struct {
	Status          *cenums.UsersToBillingPlansStatus `json:"status" gorm:"column:status;"`
	EndDate         *time.Time                        `json:"endDate" gorm:"column:end_date;"`
	NextBillingDate *time.Time                        `json:"nextBillingDate" gorm:"column:next_billing_date;"`
	FailureCount    *int32                            `json:"failureCount" gorm:"column:failure_count;"`
}

type PartialUpdateUsersToBillingPlansInput = PartialUpdateInput[UpdateUsersToBillingPlansInput]
