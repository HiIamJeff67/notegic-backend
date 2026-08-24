package other

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
)

type BadgeServiceInterface interface {
	// services for public badges
	GetPublicBadgeByPublicId(ctx context.Context, publicId uuid.UUID) (*gqlmodels.PublicBadge, *exceptions.Exception)
	GetPublicBadgeByUserPublicId(ctx context.Context, publicId uuid.UUID) (*gqlmodels.PublicBadge, *exceptions.Exception)
	GetPublicBadgesByUserPublicIds(ctx context.Context, publicIds []uuid.UUID) ([]*gqlmodels.PublicBadge, *exceptions.Exception)
}

type BadgeService struct {
	db *gorm.DB
}

func NewBadgeService(db *gorm.DB) BadgeServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &BadgeService{db: db}
}

/* ============================== Service Methods for Badge ============================== */

/* ============================== Services for GraphQL Badge ============================== */

func (s *BadgeService) GetPublicBadgeByPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*gqlmodels.PublicBadge, *exceptions.Exception) {
	db := s.db.WithContext(ctx)

	badge := schemas.Badge{}
	result := db.Table(schemas.Badge{}.TableName()).
		Where("public_id = ?", publicId).
		First(&badge)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"Badge",
			"GetPublicBadgeByPublicId",
			"Badge was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	return badge.ToPublicBadge(), nil
}

func (s *BadgeService) GetPublicBadgeByUserPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*gqlmodels.PublicBadge, *exceptions.Exception) {
	db := s.db.WithContext(ctx)

	badge := schemas.Badge{}
	result := db.Table(schemas.Badge{}.TableName()+" b").
		Select("b.*, utb.user_id").
		Joins(`LEFT JOIN "UsersToBadgesTable" utb ON utb.badge_id = b.id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = utb.user_id`).
		Where("u.public_id = ?", publicId).
		First(&badge)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"Badge",
			"GetPublicBadgeByUserPublicId",
			"Badge was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	return badge.ToPublicBadge(), nil
}

func (s *BadgeService) GetPublicBadgesByUserPublicIds(
	ctx context.Context, publicIds []uuid.UUID,
) ([]*gqlmodels.PublicBadge, *exceptions.Exception) {
	if len(publicIds) == 0 {
		return []*gqlmodels.PublicBadge{}, nil
	}

	db := s.db.WithContext(ctx)

	uniquePublicIds := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]bool)
	for _, publicId := range publicIds {
		if !seen[publicId] {
			uniquePublicIds = append(uniquePublicIds, publicId)
			seen[publicId] = true
		}
	}
	if len(uniquePublicIds) == 0 {
		return make([]*gqlmodels.PublicBadge, len(publicIds)), nil
	}

	var badgesWithPublicUserIds []*struct {
		schemas.Badge
		UserPublicId uuid.UUID `gorm:"column:user_public_id"`
	}
	result := db.Table(schemas.Badge{}.TableName()+" b").
		Select("b.*, u.public_id as user_public_id").
		Joins(`LEFT JOIN "UsersToBadgesTable" utb ON utb.badge_id = b.id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = utb.user_id`).
		Where("u.public_id IN ?", uniquePublicIds).
		Find(&badgesWithPublicUserIds)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"QueryFailed",
			"Badge",
			"GetPublicBadgesByUserPublicIds",
			"Failed to retrieve badges",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	publicIdToIndexesMap := make(map[uuid.UUID][]int)
	for index, publicId := range publicIds {
		publicIdToIndexesMap[publicId] = append(publicIdToIndexesMap[publicId], index)
	}

	publicBadges := make([]*gqlmodels.PublicBadge, len(publicIds))
	for _, badgeWithPublicUserId := range badgesWithPublicUserIds {
		for _, index := range publicIdToIndexesMap[badgeWithPublicUserId.UserPublicId] {
			publicBadges[index] = badgeWithPublicUserId.Badge.ToPublicBadge()
		}
	}

	return publicBadges, nil
}
