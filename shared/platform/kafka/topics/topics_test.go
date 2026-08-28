package topics

import "testing"

func TestAllContainsExplicitUniqueTopicSpecs(t *testing.T) {
	specifications := All()
	if len(specifications) != 8 {
		t.Fatalf("topic spec count = %d, want 8", len(specifications))
	}

	seen := make(map[string]struct{}, len(specifications))
	for _, specification := range specifications {
		if specification.Name == "" || specification.Partitions < 1 || specification.ReplicationFactor < 1 ||
			specification.Retention <= 0 || specification.CleanupPolicy == "" || specification.MinInSyncReplicas < 1 ||
			!specification.CreateDeadLetter || specification.DeadLetterRetention <= 0 {
			t.Fatalf("topic spec is incomplete: %+v", specification)
		}
		if _, exists := seen[specification.Name]; exists {
			t.Fatalf("duplicate topic spec: %q", specification.Name)
		}
		seen[specification.Name] = struct{}{}
	}
}
