package routinetask

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
)

type ResultWriteFunc func(context.Context, Result) error

type RoutineTaskCompletionPublisher func(
	context.Context,
	[]croutinetasktypes.CompletedRoutineTask,
	uuid.UUID,
) error

type ResultKind string

const (
	ResultKind_Completed ResultKind = "completed"
	ResultKind_Failed    ResultKind = "failed"
)

type Result struct {
	Kind          ResultKind
	WorkerId      uuid.UUID
	CorrelationId string
	Data          any
}

type ResultWriter struct {
	executor            routineexecution.RoutineTaskExecutionServiceInterface
	completionPublisher RoutineTaskCompletionPublisher
}

func NewResultWriter(
	executor routineexecution.RoutineTaskExecutionServiceInterface,
	completionPublisher RoutineTaskCompletionPublisher,
) *ResultWriter {
	return &ResultWriter{
		executor:            executor,
		completionPublisher: completionPublisher,
	}
}

func (w *ResultWriter) Write(
	ctx context.Context,
	result Result,
) error {
	switch result.Kind {
	case ResultKind_Completed:
		request, ok := result.Data.(cdurablejob.MarkCompletedRoutineTasksRequestDto)
		if !ok {
			return fmt.Errorf("invalid completed RoutineTask result payload %T", result.Data)
		}
		if exception := w.executor.ApplyPreparedRoutineTasks(ctx, uuid.New(), &request); exception != nil {
			return exception
		}
		if w.completionPublisher != nil {
			if err := w.completionPublisher(ctx, request.Tasks, request.WorkerId); err != nil {
				return fmt.Errorf("publish completed RoutineTask lifecycle events: %w", err)
			}
		}
		return nil
	case ResultKind_Failed:
		request, ok := result.Data.(cdurablejob.MarkFailedRoutineTasksRequestDto)
		if !ok {
			return fmt.Errorf("invalid failed RoutineTask result payload %T", result.Data)
		}
		if exception := w.executor.ApplyFailedRoutineTasks(ctx, uuid.New(), &request); exception != nil {
			return exception
		}
		return nil
	default:
		return fmt.Errorf("unsupported RoutineTask result kind %q", result.Kind)
	}
}
