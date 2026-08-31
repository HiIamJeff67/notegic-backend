package constants

import (
	"time"

	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

/* ============================== Realtime Gateway limitations ============================== */

const (
	RealtimeConnectionTicketExpiresIn   time.Duration = 5 * time.Minute
	RealtimeBlockPackTicketExpiresIn    time.Duration = 5 * time.Minute
	RealtimeProtocolVersion             int           = 1
	RealtimeWorkerProtocolVersion       int           = 1
	RealtimeMaxMessageSize              int64         = 1 << 20
	RealtimeMaxChannelsPerConnection    int           = 64
	RealtimePongWait                    time.Duration = 90 * time.Second
	RealtimePingInterval                time.Duration = 60 * time.Second
	RealtimeLeaseTTL                    time.Duration = 3 * time.Minute
	RealtimeControlWriteTimeout         time.Duration = 10 * time.Second
	RealtimeWorkerReconnectDelay        time.Duration = 2 * time.Second
	RealtimeWorkerAttachTimeout         time.Duration = 10 * time.Second
	RealtimeWorkerQueueSize             int           = 1024
	RealtimeMaxOutboundControlFrames    int           = 256
	RealtimeMaxOutboundFramesPerChannel int           = 256
	RealtimeMaxOutboundBytesPerChannel  int64         = 4 << 20
	RealtimeMaxConnectorsPerGateway     int           = 10_000
	RealtimeMaxConnectionsPerUser       int           = 8
	RealtimeRedisServerNumber           int           = 8
)

const (
	RealtimeBinaryFrameHeaderSize   int = 6
	RealtimeInternalFrameHeaderSize int = 39
)

/* ============================== Routine Task limitations ============================== */

const (
	RoutineTaskEngineMaxWorkers         int           = 8
	RoutineTaskEngineTickerDuration     time.Duration = 1 * time.Minute
	RoutineTaskExecutionLeaseTTL        time.Duration = 10 * time.Minute
	RoutineTaskClaimerMaxClaimableTasks int           = 1e6
)

/* ============================== Email Worker limitations ============================== */

const (
	EmailWorkerManagerTickerDuration time.Duration = 100 * time.Millisecond
)

/* ============================== Storage limitations ============================== */

const (
	MaxNonVideoFileSize        types.ByteType = 10 * types.MB
	MaxInMemoryStorageFileSize types.ByteType = 10 * types.MB
	MaxS3StorageFileSize       types.ByteType = 10 * types.MB
)

/* ============================== Variable constraints ============================== */

const (
	MaxUserAgentLength int = 2048
	MaxURLLength       int = 2048
	MinPasswordLength  int = 8
	MaxPasswordLength  int = 1024

	// make sure the below values are as the same as the constraint in the dto while registering or creating the user
	MaxNameLength = 32
	MinNameLength = 6
)

/* ============================== Database or orm limitation ============================== */

const (
	// the max batch size of the PostgreSQL and Gorm is limited by
	// the formula: Max Batch Size = 65535 / number of columns in the target table
	MaxBatchCreateBlockSize int = 3000
)

const (
	DefaultSearchLimit = 10
	MaxSearchLimit     = 100
)
