package blockpackyjstriggersql

import (
	_ "embed"
)

const SyncBlockPackYjsDocumentDeletedAtTriggerFunctionName = "trigger_function_sync_block_pack_yjs_document_deleted_at"

var (
	//go:embed sync_block_pack_yjs_document_deleted_at_trigger.sql
	SyncBlockPackYjsDocumentDeletedAtTriggerSQL string
)
