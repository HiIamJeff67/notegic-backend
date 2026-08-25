package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UsersToBillingPlansStatus string

const (
	UsersToBillingPlansStatus_ApprovalPending UsersToBillingPlansStatus = "APPROVAL_PENDING"
	UsersToBillingPlansStatus_Approved        UsersToBillingPlansStatus = "APPROVED"
	UsersToBillingPlansStatus_Active          UsersToBillingPlansStatus = "ACTIVE"
	UsersToBillingPlansStatus_Suspended       UsersToBillingPlansStatus = "SUSPENDED"
	UsersToBillingPlansStatus_Cancelled       UsersToBillingPlansStatus = "CANCELLED"
	UsersToBillingPlansStatus_Expired         UsersToBillingPlansStatus = "EXPIRED"
)

var AllUsersToBillingPlansStatuses = []UsersToBillingPlansStatus{
	UsersToBillingPlansStatus_ApprovalPending,
	UsersToBillingPlansStatus_Approved,
	UsersToBillingPlansStatus_Active,
	UsersToBillingPlansStatus_Suspended,
	UsersToBillingPlansStatus_Cancelled,
	UsersToBillingPlansStatus_Expired,
}

var AllUsersToBillingPlansStatusStrings = []string{
	string(UsersToBillingPlansStatus_ApprovalPending),
	string(UsersToBillingPlansStatus_Approved),
	string(UsersToBillingPlansStatus_Active),
	string(UsersToBillingPlansStatus_Suspended),
	string(UsersToBillingPlansStatus_Cancelled),
	string(UsersToBillingPlansStatus_Expired),
}

func (value UsersToBillingPlansStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UsersToBillingPlansStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UsersToBillingPlansStatus(string(v))
		return nil
	case string:
		*value = UsersToBillingPlansStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UsersToBillingPlansStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UsersToBillingPlansStatus) String() string {
	return string(value)
}

func (value *UsersToBillingPlansStatus) IsValidEnum() bool {
	return slices.Contains(AllUsersToBillingPlansStatuses, *value)
}

func ConvertStringToUsersToBillingPlansStatus(enumString string) (*UsersToBillingPlansStatus, error) {
	for _, enumValue := range AllUsersToBillingPlansStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UsersToBillingPlansStatus: %s", enumString)
}
