package routinetask

import (
	"testing"

	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
)

func TestNewManagerInjectsPlanService(t *testing.T) {
	planService := routineexecution.NewPlanService(nil, nil)
	manager := NewManager(planService, nil, nil, nil)

	if manager.planService != planService {
		t.Fatal("routine task plan service was not injected")
	}
}
