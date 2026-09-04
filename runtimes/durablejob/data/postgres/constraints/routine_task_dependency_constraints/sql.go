package routinetaskdependencyconstraints

import (
	_ "embed"
)

//go:embed routine_task_dependency_foreign_keys.sql
var RoutineTaskDependencyForeignKeysSQL string
