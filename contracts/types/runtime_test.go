package types

import "testing"

func TestRuntimeValues(t *testing.T) {
	for _, runtime := range AllRuntimes {
		if !runtime.IsValid() {
			t.Fatalf("runtime %q should be valid", runtime)
		}
	}

	if runtime := Runtime("unknown"); runtime.IsValid() {
		t.Fatal("unknown runtime should be invalid")
	}

	runtime, err := ConvertStringToRuntime("durablejob")
	if err != nil || runtime == nil || *runtime != Runtime_DurableJob {
		t.Fatalf("ConvertStringToRuntime() = %v, %v", runtime, err)
	}
}
