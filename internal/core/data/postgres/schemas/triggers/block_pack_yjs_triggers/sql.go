package blockpackyjstriggersql

import (
	_ "embed"
)

var (
	//go:embed sync_block_pack_yjs_document_deleted_at_trigger.sql
	SyncBlockPackYjsDocumentDeletedAtTriggerSQL string
)
