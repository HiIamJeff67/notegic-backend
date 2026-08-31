package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type Handler struct{}

func (Handler) BuildGetResult(
	itemId uuid.UUID,
	found bool,
	value any,
) (croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	result := croutinetasktypes.ExecutionResult{At: time.Now().UTC()}
	item := croutinetasktypes.ExecutionItemResult{ItemId: itemId.String()}
	if !found {
		item.Status = croutinetasktypes.ExecutionItemStatus_Skipped
		item.Reason = "object_not_found_or_forbidden"
		result.Skipped = 1
		result.Items = []croutinetasktypes.ExecutionItemResult{item}
		return result, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return result, cexceptions.New(
			"FailedToEncode",
			"RoutineTask",
			"Get",
			"Failed to encode the retrieved object",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	item.Status = croutinetasktypes.ExecutionItemStatus_Retrieved
	item.Data = data
	result.Retrieved = 1
	result.Items = []croutinetasktypes.ExecutionItemResult{item}
	return result, nil
}
