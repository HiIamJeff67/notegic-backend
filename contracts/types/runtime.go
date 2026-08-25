package types

import (
	"fmt"
	"slices"
)

// Runtime identifies a Notegic runtime across shared platform boundaries.
type Runtime string

const (
	Runtime_Core         Runtime = "core"
	Runtime_DurableJob   Runtime = "durablejob"
	Runtime_Notification Runtime = "notification"
)

var AllRuntimes = []Runtime{
	Runtime_Core,
	Runtime_DurableJob,
	Runtime_Notification,
}

var AllRuntimeStrings = []string{
	string(Runtime_Core),
	string(Runtime_DurableJob),
	string(Runtime_Notification),
}

func (value Runtime) String() string {
	return string(value)
}

func (value *Runtime) IsValid() bool {
	return value != nil && slices.Contains(AllRuntimes, *value)
}

func ConvertStringToRuntime(runtimeString string) (*Runtime, error) {
	for _, runtime := range AllRuntimes {
		if string(runtime) == runtimeString {
			return &runtime, nil
		}
	}
	return nil, fmt.Errorf("invalid runtime: %s", runtimeString)
}
