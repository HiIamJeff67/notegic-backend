package subshelfconstraints

import (
	"strings"
	"testing"
)

func TestSubShelfPreviousReferenceConstraintIsDeferred(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (prev_sub_shelf_id)",
		"REFERENCES \"SubShelfTable\" (id)",
		"DEFERRABLE INITIALLY DEFERRED",
	} {
		if !strings.Contains(SubShelfPreviousReferenceConstraintSQL, fragment) {
			t.Fatalf("SubShelfPreviousReferenceConstraintSQL must contain %q", fragment)
		}
	}
}
