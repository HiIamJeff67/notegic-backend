package inputs

type SynchronizeRateLimitRecordCacheInput struct {
	NumOfChangingTokens int32 `json:"numOfChangingTokens"`
	IsAccumulated       bool  `json:"isAccumulated"`
}

type BatchSynchronizeRateLimitRecordCacheInput struct {
	Identifier string                               `json:"identifier"`
	Input      SynchronizeRateLimitRecordCacheInput `json:"input"`
}
