package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Theme struct {
	Id            uuid.UUID `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	PublicId      uuid.UUID `json:"publicId" gorm:"column:public_id; type:uuid; unique; not null; default:gen_random_uuid();"`
	AuthorId      uuid.UUID `json:"authorId" gorm:"column:author_id; type:uuid; not null; uniqueIndex;"`
	Name          string    `json:"name" gorm:"column:name; size:128; unique; not null;"`
	IsDark        bool      `json:"isDark" gorm:"column:is_dark; type:boolean; not null; default:true;"`
	Version       string    `json:"version" gorm:"column:version; not null;"`
	IsDefault     bool      `json:"isDefault" gorm:"column:is_default; type:boolean; not null;"`
	DownloadURL   *string   `json:"downloadURL" gorm:"column:download_url;"`
	DownloadCount int64     `json:"downloadCount" gorm:"column:download_count; type:bigint; not null default:0;"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`

	// relation
	Author User `json:"author" gorm:"foreignKey:AuthorId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Theme Table Name
func (Theme) TableName() string {
	return "ThemeTable"
}

// Theme Table Relations
type ThemeRelation platformpostgres.RelationName

const (
	ThemeRelation_Author ThemeRelation = "Themes"
)

/* ============================== Relative Type Conversion ============================== */

func (t *Theme) ToPublicTheme() *gqlmodels.PublicTheme {
	return &gqlmodels.PublicTheme{
		PublicID:      t.PublicId,
		Name:          t.Name,
		IsDark:        t.IsDark,
		Version:       t.Version,
		IsDefault:     t.IsDefault,
		DownloadURL:   t.DownloadURL,
		DownloadCount: t.DownloadCount,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		AuthorID:      t.AuthorId,
	}
}

/* ============================== Trigger Hook ============================== */

func (t *Theme) BeforeCreate(tx *gorm.DB) error {
	if t.PublicId == uuid.Nil {
		t.PublicId = uuid.New()
	}
	return nil
}
