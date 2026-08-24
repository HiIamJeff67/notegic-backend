package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type Badge struct {
	Id          uuid.UUID       `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid()"`
	PublicId    uuid.UUID       `json:"publicId" gorm:"column:public_id; type:uuid; unique; not null; default:gen_random_uuid();"`
	Title       string          `json:"title" gorm:"column:title; not null; size:64;"`
	Description string          `json:"description" gorm:"column:description; not null; size:256;"`
	Type        enums.BadgeType `json:"type" gorm:"column:type; type:\"BadgeType\"; not null; default:'Bronze';"`
	ImageURL    *string         `json:"imageURL" gorm:"column:image_url;"`
	CreatedAt   time.Time       `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relation
	UsersToBadges []UsersToBadges `json:"usersToBadges" gorm:"foreignKey:BadgeId;"`
}

// Badge Table Name
func (Badge) TableName() string {
	return "BadgeTable"
}

// Badge Table Relations
type BadgeRelation platformpostgres.RelationName

const (
	BadgeRelation_UsersToBadges BadgeRelation = "UsersToBadges"
)

/* ============================== Relative Type Conversions ============================== */

func (b *Badge) ToPublicBadge() *gqlmodels.PublicBadge {
	return &gqlmodels.PublicBadge{
		PublicID:    b.PublicId,
		Title:       b.Title,
		Description: b.Description,
		Type:        *b.Type.ToContractable(),
		ImageURL:    b.ImageURL,
		CreatedAt:   b.CreatedAt,
	}
}

/* ============================== Trigger Hook ============================== */

func (b *Badge) BeforeCreate(tx *gorm.DB) error {
	if b.PublicId == uuid.Nil {
		b.PublicId = uuid.New()
	}
	return nil
}
