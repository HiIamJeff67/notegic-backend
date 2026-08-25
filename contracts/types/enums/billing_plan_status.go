package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type BillingPlanStatus string

const (
	BillingPlanStatus_Created  BillingPlanStatus = "CREATED"
	BillingPlanStatus_Active   BillingPlanStatus = "ACTIVE"
	BillingPlanStatus_Inactive BillingPlanStatus = "INACTIVE"
)

var AllBillingPlanStatuses = []BillingPlanStatus{
	BillingPlanStatus_Created,
	BillingPlanStatus_Active,
	BillingPlanStatus_Inactive,
}

var AllBillingPlanStatusStrings = []string{
	string(BillingPlanStatus_Created),
	string(BillingPlanStatus_Active),
	string(BillingPlanStatus_Inactive),
}

func (value BillingPlanStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *BillingPlanStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = BillingPlanStatus(string(v))
		return nil
	case string:
		*value = BillingPlanStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value BillingPlanStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value BillingPlanStatus) String() string {
	return string(value)
}

func (value *BillingPlanStatus) IsValidEnum() bool {
	return slices.Contains(AllBillingPlanStatuses, *value)
}

func ConvertStringToBillingPlanStatus(enumString string) (*BillingPlanStatus, error) {
	for _, enumValue := range AllBillingPlanStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid BillingPlanStatus: %s", enumString)
}
