package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UsersToBillingPlansStatus enumcontract.UsersToBillingPlansStatus

func (value *UsersToBillingPlansStatus) ToContractable() *enumcontract.UsersToBillingPlansStatus {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UsersToBillingPlansStatus(*value)
	return &contractValue
}

func UsersToBillingPlansStatusToStorable(value *enumcontract.UsersToBillingPlansStatus) *UsersToBillingPlansStatus {
	if value == nil {
		return nil
	}

	storableValue := UsersToBillingPlansStatus(*value)
	return &storableValue
}

const (
	UsersToBillingPlansStatus_ApprovalPending UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_ApprovalPending)
	UsersToBillingPlansStatus_Approved        UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_Approved)
	UsersToBillingPlansStatus_Active          UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_Active)
	UsersToBillingPlansStatus_Suspended       UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_Suspended)
	UsersToBillingPlansStatus_Cancelled       UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_Cancelled)
	UsersToBillingPlansStatus_Expired         UsersToBillingPlansStatus = UsersToBillingPlansStatus(enumcontract.UsersToBillingPlansStatus_Expired)
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

func (utbps UsersToBillingPlansStatus) Name() string {
	return reflect.TypeOf(utbps).Name()
}

func (utbps *UsersToBillingPlansStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*utbps = UsersToBillingPlansStatus(string(v))
		return nil
	case string:
		*utbps = UsersToBillingPlansStatus(v)
		return nil
	}
	return scanError(value, utbps)
}

func (utbps UsersToBillingPlansStatus) Value() (driver.Value, error) {
	return string(utbps), nil
}

func (utbps UsersToBillingPlansStatus) String() string {
	return string(utbps)
}

func (utbps *UsersToBillingPlansStatus) IsValidEnum() bool {
	return slices.Contains(AllUsersToBillingPlansStatuses, *utbps)
}

func ConvertStringToUsersToBillingPlansStatus(enumString string) (*UsersToBillingPlansStatus, error) {
	for _, supportedCurrencyCode := range AllUsersToBillingPlansStatuses {
		if string(supportedCurrencyCode) == enumString {
			return &supportedCurrencyCode, nil
		}
	}
	return nil, fmt.Errorf("invalid users to billing plans status: %s", enumString)
}
