package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UserViewByPublicId(publicId uuid.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`"public_id" = ?`, publicId)
	}
}
