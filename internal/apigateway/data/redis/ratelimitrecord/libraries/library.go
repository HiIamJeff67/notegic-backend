package libraries

import _ "embed"

const (
	RateLimitRecordLibrary = "rate_limit_record_library"

	BatchSynchronizeRateLimitRecordByFormattedKeysFunction = "batch_synchronize_rate_limit_record_by_formatted_keys"
	BatchDeleteRateLimitRecordByFormattedKeysFunction      = "batch_delete_rate_limit_record_by_formatted_keys"
)

//go:embed rate_limit_record_library.lua
var RateLimitRecordLibraryContent string
