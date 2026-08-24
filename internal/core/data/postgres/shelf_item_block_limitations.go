package postgres

import types "github.com/HiIamJeff67/notegic-backend/shared/types"

const (
	MaxSubShelvesOfRootShelf int32 = 1e+2
	MaxContentOfRootShelf    int32 = 1e+2
	MaxMaterialsOfRootShelf  int32 = 1e+2
	MaxBlockPackOfRootShelf  int32 = 1e+2

	MaxSubShelvesOfSubShelf int32 = 1e+2
	MaxContentOfSubShelf    int32 = 1e+2
	MaxMaterialsOfSubShelf  int32 = 1e+2
	MaxBlockPackOfSubShelf  int32 = 1e+2

	PeekFileSize             types.ByteType = 256 * types.Byte
	MaxMaterialTextFileSize  types.ByteType = 5 * types.MB
	MaxMaterialImageFileSize types.ByteType = 20 * types.MB
	MaxMaterialVideoFileSize types.ByteType = 100 * types.MB
	MaxMaterialAudioFileSize types.ByteType = 20 * types.MB
)
