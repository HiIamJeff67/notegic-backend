package schemas

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// PlanLimitation is the shared persistence model for plan limits. It is
// writable only by administrative flows, but readable by runtimes that need
// to enforce a user's plan constraints.
type PlanLimitation struct {
	Key                            enums.UserPlan `json:"key" gorm:"column:key; type:\"UserPlan\"; primaryKey;"`
	MaxRootShelfCount              int32          `json:"maxRootShelfCount" gorm:"column:max_root_shelf_count; type:integer; not null;"`
	MaxBlockPackCount              int32          `json:"maxBlockPackCount" gorm:"column:max_block_pack_count; type:integer; not null;"`
	MaxBlockCount                  int32          `json:"maxBlockCount" gorm:"column:max_block_count; type:integer; not null;"`
	MaxMaterialCount               int32          `json:"maxMaterialCount" gorm:"column:max_material_count; type:integer; not null;"`
	MaxWorkflowCount               int32          `json:"maxWorkflowCount" gorm:"column:max_work_flow_count; type:integer; not null;"`
	MaxAdditionalItemCount         int32          `json:"maxAdditionalItemCount" gorm:"column:max_additional_item_count; type:integer; not null;"`
	MaxSubShelfCountPerRootShelf   int32          `json:"maxSubShelfCountPerRootShelf" gorm:"column:max_sub_shelf_count_per_root_shelf; type:integer; not null;"`
	MaxItemCountPerRootShelf       int32          `json:"maxItemCountPerRootShelf" gorm:"column:max_item_count_per_root_shelf; type:integer; not null;"`
	MaxBlockCountPerBlockPack      int32          `json:"maxBlockCountPerBlockPack" gorm:"column:max_block_count_per_block_pack; type:integer; not null;"`
	MaxMaterialSize                int64          `json:"maxMaterialSize" gorm:"column:max_material_size; type:bigint; not null;"`
	MaxStationCount                int32          `json:"maxStationCount" gorm:"column:max_station_count; type:integer; not null;"`
	MaxRoutineTagCount             int32          `json:"maxRoutineTagCount" gorm:"column:max_routine_tag_count; type:integer; not null;"`
	MaxRoutineCountPerStation      int32          `json:"maxRoutineCountPerStation" gorm:"column:max_routine_count_per_station; type:integer; not null;"`
	MaxRoutineTaskCostUnitCount    int32          `json:"maxRoutineTaskCostUnitCount" gorm:"column:max_routine_task_cost_unit_count; type:integer; not null;"`
	MaxRoutineTaskAttempts         int32          `json:"maxRoutineTaskAttempts" gorm:"column:max_routine_task_attempts; type:integer; not null;"`
	MaxRealtimeRoomSubscriberCount int32          `json:"maxRealtimeRoomSubscriberCount" gorm:"column:max_realtime_room_subscriber_count; type:integer; not null; default:0;"`
	UpdatedAt                      time.Time      `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt                      time.Time      `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`
}

func (PlanLimitation) TableName() string {
	return platformpostgres.TableName_PlanLimitationTable.String()
}
