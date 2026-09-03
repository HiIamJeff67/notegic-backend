package routinetask

import (
	"testing"

	durablejobexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/exceptions"
	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
)

func TestNewManagerInjectsPlanService(t *testing.T) {
	planService := routineexecution.NewPlanService(nil, nil, durablejobexceptions.NewRoutineTaskException())
	manager := NewManager(nil, planService, nil, nil, nil)

	if manager.planService != planService {
		t.Fatal("routine task plan service was not injected")
	}
}
