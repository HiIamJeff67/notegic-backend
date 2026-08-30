package realtimetypes

type InternalFrameType byte

const (
	InternalFrameType_Attach                       InternalFrameType = 1
	InternalFrameType_Detach                       InternalFrameType = 2
	InternalFrameType_YjsDocument                  InternalFrameType = 3
	InternalFrameType_Awareness                    InternalFrameType = 4
	InternalFrameType_ResyncRequired               InternalFrameType = 5
	InternalFrameType_PermissionRevoked            InternalFrameType = 6
	InternalFrameType_LoadYjsDocument              InternalFrameType = 7
	InternalFrameType_YjsDocumentLoaded            InternalFrameType = 8
	InternalFrameType_AppendYjsUpdate              InternalFrameType = 9
	InternalFrameType_YjsUpdatePersisted           InternalFrameType = 10
	InternalFrameType_YjsPersistenceFailed         InternalFrameType = 11
	InternalFrameType_ApplyBlockProjection         InternalFrameType = 12
	InternalFrameType_BlockProjectionApplied       InternalFrameType = 13
	InternalFrameType_BlockProjectionFailed        InternalFrameType = 14
	InternalFrameType_AppendYjsUpdateBatch         InternalFrameType = 15
	InternalFrameType_LoadCompactableYjsDocument   InternalFrameType = 16
	InternalFrameType_CompactableYjsDocumentLoaded InternalFrameType = 17
	InternalFrameType_ApplyCompactedYjsDocument    InternalFrameType = 18
	InternalFrameType_YjsDocumentCompacted         InternalFrameType = 19
	InternalFrameType_YjsDocumentCompactionFailed  InternalFrameType = 20
	InternalFrameType_Attached                     InternalFrameType = 21
	InternalFrameType_BlockPackQuotaExceeded       InternalFrameType = 22
)
