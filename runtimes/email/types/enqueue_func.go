package types

type EnqueueFunc func(
	emailObject EmailObject,
	taskType EmailTaskType,
	maxRetries int,
	priority int,
) error
