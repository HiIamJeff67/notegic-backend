package apicontract

const (
	GetMyRoutineTaskByIdOperation               = "routine-task.get-by-id"
	GetAllMyRoutineTasksByRoutineIdsOperation   = "routine-task.get-all-by-routine-ids"
	GetAllMyRoutineTasksOperation               = "routine-task.get-all"
	CreateRoutineTaskByRoutineIdOperation       = "routine-task.create-by-routine-id"
	UpdateMyRoutineTaskByIdOperation            = "routine-task.update"
	HardDeleteMyRoutineTaskByIdOperation        = "routine-task.hard-delete"
	HardDeleteMyRoutineTasksByIdsOperation      = "routine-task.hard-delete-many"
	VisualizeMyRoutineTaskPurposeCountOperation = "routine-task.visualize-purpose-count"
	SearchRoutineTasksOperation                 = "graphql.search-routine-tasks"
)
