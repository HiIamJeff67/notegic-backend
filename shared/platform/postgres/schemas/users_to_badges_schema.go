package schemas

import (
	"time"

	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type UsersToBadges struct {
	UserId    uuid.UUID `json:"userId" gorm:"column:user_id; type:uuid; primaryKey;"`
	BadgeId   uuid.UUID `json:"badgeId" gorm:"column:badge_id; type:uuid; primaryKey;"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	User  User  `gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Badge Badge `gorm:"foreignKey:BadgeId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// UsersToBadges Table Name
func (UsersToBadges) TableName() string {
	return postgres.TableName_UsersToBadgesTable.String()
}

// UsersToBadges Table Relations
type UsersToBadgesRelation postgres.RelationName

const (
	UsersToBadgesRelation_User  UsersToBadgesRelation = "User"
	UsersToBadgesRelation_Badge UsersToBadgesRelation = "Badge"
)
