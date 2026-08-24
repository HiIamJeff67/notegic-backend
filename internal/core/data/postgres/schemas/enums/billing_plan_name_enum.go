package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type BillingPlanName enumcontract.BillingPlanName

func (value *BillingPlanName) ToContractable() *enumcontract.BillingPlanName {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.BillingPlanName(*value)
	return &contractValue
}

func BillingPlanNameToStorable(value *enumcontract.BillingPlanName) *BillingPlanName {
	if value == nil {
		return nil
	}

	storableValue := BillingPlanName(*value)
	return &storableValue
}

const (
	BillingPlanName_NotegicMonthlyFreePlan       BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicMonthlyFreePlan)
	BillingPlanName_NotegicMonthlyProPlan        BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicMonthlyProPlan)
	BillingPlanName_NotegicYearlyProPlan         BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicYearlyProPlan)
	BillingPlanName_NotegicMonthlyPremiumPlan    BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicMonthlyPremiumPlan)
	BillingPlanName_NotegicYearlyPremiumPlan     BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicYearlyPremiumPlan)
	BillingPlanName_NotegicMonthlyUltimatePlan   BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicMonthlyUltimatePlan)
	BillingPlanName_NotegicYearlyUltimatePlan    BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicYearlyUltimatePlan)
	BillingPlanName_NotegicMonthlyEnterprisePlan BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicMonthlyEnterprisePlan)
	BillingPlanName_NotegicYearlyEnterprisePlan  BillingPlanName = BillingPlanName(enumcontract.BillingPlanName_NotegicYearlyEnterprisePlan)
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

func (bpn BillingPlanName) Name() string {
	return reflect.TypeOf(bpn).Name()
}

func (bpn *BillingPlanName) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*bpn = BillingPlanName(string(v))
		return nil
	case string:
		*bpn = BillingPlanName(v)
		return nil
	}
	return scanError(value, bpn)
}

func (bpn BillingPlanName) Value() (driver.Value, error) {
	return string(bpn), nil
}

func (bpn BillingPlanName) String() string {
	return string(bpn)
}

func (bpn *BillingPlanName) IsValidEnum() bool {
	return slices.Contains(AllBillingPlanNames, *bpn)
}

func ConvertStringToBillingPlanName(enumString string) (*BillingPlanName, error) {
	for _, billingPlanName := range AllBillingPlanNames {
		if string(billingPlanName) == enumString {
			return &billingPlanName, nil
		}
	}
	return nil, fmt.Errorf("invalid billing plan name: %s", enumString)
}
