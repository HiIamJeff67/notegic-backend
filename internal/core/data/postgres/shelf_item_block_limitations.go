package postgres

import stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

const (
	MaxSubShelvesOfRootShelf int32 = 1e+2
	MaxContentOfRootShelf    int32 = 1e+2
	MaxMaterialsOfRootShelf  int32 = 1e+2
	MaxBlockPackOfRootShelf  int32 = 1e+2

	MaxSubShelvesOfSubShelf int32 = 1e+2
	MaxContentOfSubShelf    int32 = 1e+2
	MaxMaterialsOfSubShelf  int32 = 1e+2
	MaxBlockPackOfSubShelf  int32 = 1e+2

	PeekFileSize             stypes.ByteType = 256 * stypes.Byte
	MaxMaterialTextFileSize  stypes.ByteType = 5 * stypes.MB
	MaxMaterialImageFileSize stypes.ByteType = 20 * stypes.MB
	MaxMaterialVideoFileSize stypes.ByteType = 100 * stypes.MB
	MaxMaterialAudioFileSize stypes.ByteType = 20 * stypes.MB
)
