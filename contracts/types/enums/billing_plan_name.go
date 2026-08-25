package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type BillingPlanName string

const (
	BillingPlanName_NotegicMonthlyFreePlan       BillingPlanName = "Notegic Monthly Free Plan"
	BillingPlanName_NotegicMonthlyProPlan        BillingPlanName = "Notegic Monthly Pro Plan"
	BillingPlanName_NotegicYearlyProPlan         BillingPlanName = "Notegic Yearly Pro Plan"
	BillingPlanName_NotegicMonthlyPremiumPlan    BillingPlanName = "Notegic Monthly Premium Plan"
	BillingPlanName_NotegicYearlyPremiumPlan     BillingPlanName = "Notegic Yearly Premium Plan"
	BillingPlanName_NotegicMonthlyUltimatePlan   BillingPlanName = "Notegic Monthly Ultimate Plan"
	BillingPlanName_NotegicYearlyUltimatePlan    BillingPlanName = "Notegic Yearly Ultimate Plan"
	BillingPlanName_NotegicMonthlyEnterprisePlan BillingPlanName = "Notegic Monthly Enterprise Plan"
	BillingPlanName_NotegicYearlyEnterprisePlan  BillingPlanName = "Notegic Yearly Enterprise Plan"
)

var AllBillingPlanNames = []BillingPlanName{
	BillingPlanName_NotegicMonthlyFreePlan,
	BillingPlanName_NotegicMonthlyProPlan,
	BillingPlanName_NotegicYearlyProPlan,
	BillingPlanName_NotegicMonthlyPremiumPlan,
	BillingPlanName_NotegicYearlyPremiumPlan,
	BillingPlanName_NotegicMonthlyUltimatePlan,
	BillingPlanName_NotegicYearlyUltimatePlan,
	BillingPlanName_NotegicMonthlyEnterprisePlan,
	BillingPlanName_NotegicYearlyEnterprisePlan,
}

var AllBillingPlanNameStrings = []string{
	string(BillingPlanName_NotegicMonthlyFreePlan),
	string(BillingPlanName_NotegicMonthlyProPlan),
	string(BillingPlanName_NotegicYearlyProPlan),
	string(BillingPlanName_NotegicMonthlyPremiumPlan),
	string(BillingPlanName_NotegicYearlyPremiumPlan),
	string(BillingPlanName_NotegicMonthlyUltimatePlan),
	string(BillingPlanName_NotegicYearlyUltimatePlan),
	string(BillingPlanName_NotegicMonthlyEnterprisePlan),
	string(BillingPlanName_NotegicYearlyEnterprisePlan),
}

func (value BillingPlanName) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *BillingPlanName) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = BillingPlanName(string(v))
		return nil
	case string:
		*value = BillingPlanName(v)
		return nil
	}
	return scanError(raw, value)
}

func (value BillingPlanName) Value() (driver.Value, error) {
	return string(value), nil
}

func (value BillingPlanName) String() string {
	return string(value)
}

func (value *BillingPlanName) IsValidEnum() bool {
	return slices.Contains(AllBillingPlanNames, *value)
}

func ConvertStringToBillingPlanName(enumString string) (*BillingPlanName, error) {
	for _, enumValue := range AllBillingPlanNames {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid BillingPlanName: %s", enumString)
}
