package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutinePhase string

const (
	RoutinePhase_Claimed     RoutinePhase = "Claimed"
	RoutinePhase_Preparation RoutinePhase = "Preparation"
	RoutinePhase_Plan        RoutinePhase = "Plan"
	RoutinePhase_Execution   RoutinePhase = "Execution"
	RoutinePhase_Recovery    RoutinePhase = "Recovery"
	RoutinePhase_Analysis    RoutinePhase = "Analysis"
)

var AllRoutinePhases = []RoutinePhase{
	RoutinePhase_Claimed,
	RoutinePhase_Preparation,
	RoutinePhase_Plan,
	RoutinePhase_Execution,
	RoutinePhase_Recovery,
	RoutinePhase_Analysis,
}

var AllRoutinePhaseStrings = []string{
	string(RoutinePhase_Claimed),
	string(RoutinePhase_Preparation),
	string(RoutinePhase_Plan),
	string(RoutinePhase_Execution),
	string(RoutinePhase_Recovery),
	string(RoutinePhase_Analysis),
}

func (value RoutinePhase) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutinePhase) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutinePhase(string(v))
		return nil
	case string:
		*value = RoutinePhase(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutinePhase) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutinePhase) String() string {
	return string(value)
}

func (value *RoutinePhase) IsValidEnum() bool {
	return slices.Contains(AllRoutinePhases, *value)
}

func ConvertStringToRoutinePhase(enumString string) (*RoutinePhase, error) {
	for _, enumValue := range AllRoutinePhases {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutinePhase: %s", enumString)
}
