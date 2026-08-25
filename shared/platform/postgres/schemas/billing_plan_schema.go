package schemas

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// This table is only mutatable by the admin, and accessable by both client user and admin.
// To declare the value or data of this table, you MUST use the seeding method under notegic-backend/internal/core/data/postgres/seeds/
type BillingPlan struct {
	Id           string                      `json:"id" gorm:"column:id; primaryKey;"`
	ProductId    string                      `json:"productId" gorm:"column:product_id; not null;"`
	Name         enums.BillingPlanName       `json:"name" gorm:"column:name; type:\"BillingPlanName\"; unique; not null;"`
	Status       enums.BillingPlanStatus     `json:"status" gorm:"column:status; type:\"BillingPlanStatus\"; not null;"`
	IntervalUnit enums.BillingIntervalUnit   `json:"intervalUnit" gorm:"column:interval_unit; type:\"BillingIntervalUnit\"; not null;"`
	Price        float64                     `json:"price" gorm:"column:price; not null;"`
	CurrencyCode enums.SupportedCurrencyCode `json:"currencyCode" gorm:"column:currency_code; type:\"SupportedCurrencyCode\"; not null;"`
	UpdatedAt    time.Time                   `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt    time.Time                   `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`
}

// Plan Limitation Table Name
func (BillingPlan) TableName() string {
	return platformpostgres.TableName_BillingPlanTable.String()
}
