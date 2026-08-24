package schemas

import (
	"time"

	"github.com/google/uuid"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

// the structure of subscriptions or checkouts or anything that related to the payment and the user
type UsersToBillingPlans struct {
	Id              uuid.UUID                       `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	UserId          uuid.UUID                       `json:"userId" gorm:"column:user_id; type:uuid; not null; uniqueIndex:users_to_billing_plans_idx_user_id_billing_plan_id_partial_status;"`
	BillingPlanId   string                          `json:"billingPlanId" gorm:"column:billing_plan_id; not null; uniqueIndex:users_to_billing_plans_idx_user_id_billing_plan_id_partial_active;"`
	Status          enums.UsersToBillingPlansStatus `json:"status" gorm:"column:status; type:\"UsersToBillingPlansStatus\"; not null;"`
	StartDate       time.Time                       `json:"startTime" gorm:"column:start_date; type:timestamptz; not null; default:NOW();"`
	EndDate         *time.Time                      `json:"endDate" gorm:"column:end_date; type:timestamptz; default:null;"`
	NextBillingDate time.Time                       `json:"nextBillingDate" gorm:"column:next_billing_date; type:timestamptz;"`
	FailureCount    int32                           `json:"failureCount" gorm:"column:failure_count; type:integer; not null; default:0;"`
	UpdatedAt       time.Time                       `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt       time.Time                       `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	User        User        `gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	BillingPlan BillingPlan `gorm:"foreignKey:BillingPlanId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// UsersToBillingPlans Table Name
func (UsersToBillingPlans) TableName() string {
	return "UsersToBillingPlansTable"
}

// UsersToBilling Plans Relations
type UsersToBullingPlansRelation platformpostgres.RelationName

const (
	UsersToBillingPlansRelation_User        UsersToBullingPlansRelation = "User"
	UsersToBullingPlansRelation_BillingPlan UsersToBullingPlansRelation = "BillingPlan"
)
