package routinetask

import (
	_ "embed"
)

//go:embed routine_record_initial_aggregate_update.sql
var UpdateRoutineRecordInitialAggregateSQL string

//go:embed routine_record_aggregate_update.sql
var UpdateRoutineRecordAggregateSQL string

//go:embed routine_record_snapshot_update.sql
var UpdateRoutineRecordSnapshotSQL string

//go:embed routine_task_record_result_snapshot_update.sql
var UpdateRoutineTaskRecordResultSnapshotSQL string

//go:embed routine_task_record_dependency_block.sql
var BlockRoutineTaskRecordDependenciesSQL string
