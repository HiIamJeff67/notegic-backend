package realtimetypes

type YjsPersistenceFailureType byte

const (
	YjsPersistenceFailureType_Retryable YjsPersistenceFailureType = 1
	YjsPersistenceFailureType_Terminal  YjsPersistenceFailureType = 2
)
