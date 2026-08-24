package schemas

import (
	"time"

	"github.com/google/uuid"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// APIKey stores only the digest of a credential. The clear-text secret is
// returned once at creation time and must never be persisted or logged.
type APIKey struct {
	Id         uuid.UUID  `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	PublicId   uuid.UUID  `json:"publicId" gorm:"column:public_id; type:uuid; unique; not null; default:gen_random_uuid();"`
	UserId     uuid.UUID  `json:"userId" gorm:"column:user_id; type:uuid; not null; index;"`
	Name       string     `json:"name" gorm:"column:name; not null; size:64;"`
	KeyPrefix  string     `json:"keyPrefix" gorm:"column:key_prefix; not null; size:16; index;"`
	KeyHash    string     `json:"-" gorm:"column:key_hash; not null; size:64; unique;"`
	LastUsedAt *time.Time `json:"lastUsedAt" gorm:"column:last_used_at; type:timestamptz;"`
	ExpiresAt  *time.Time `json:"expiresAt" gorm:"column:expires_at; type:timestamptz;"`
	RevokedAt  *time.Time `json:"revokedAt" gorm:"column:revoked_at; type:timestamptz; index;"`
	CreatedAt  time.Time  `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`
	UpdatedAt  time.Time  `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
}

func (APIKey) TableName() string {
	return "APIKeyTable"
}

type APIKeyRelation platformpostgres.RelationName
