package other

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type BadgeServiceInterface interface {
	/* ============================== GraphQL Methods ============================== */
	GetPublicBadgeByPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicBadge, *cexceptions.Exception)
	GetPublicBadgeByUserPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicBadge, *cexceptions.Exception)
	GetPublicBadgesByUserPublicIds(ctx context.Context, publicIds []uuid.UUID) ([]*cgqlmodels.PublicBadge, *cexceptions.Exception)
}

type BadgeService struct {
	db *gorm.DB
}

func NewBadgeService(db *gorm.DB) BadgeServiceInterface {
	return &BadgeService{db: db}
}

/* ============================== GraphQL Methods ============================== */

func (s *BadgeService) GetPublicBadgeByPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*cgqlmodels.PublicBadge, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	badge := sschemas.Badge{}
	result := db.Table(sschemas.Badge{}.TableName()).
		Where("public_id = ?", publicId).
		First(&badge)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
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
) (*cgqlmodels.PublicBadge, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	badge := sschemas.Badge{}
	result := db.Table(sschemas.Badge{}.TableName()+" b").
		Select("b.*, utb.user_id").
		Joins(`LEFT JOIN "UsersToBadgesTable" utb ON utb.badge_id = b.id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = utb.user_id`).
		Where("u.public_id = ?", publicId).
		First(&badge)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
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
) ([]*cgqlmodels.PublicBadge, *cexceptions.Exception) {
	if len(publicIds) == 0 {
		return []*cgqlmodels.PublicBadge{}, nil
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
		return make([]*cgqlmodels.PublicBadge, len(publicIds)), nil
	}

	var badgesWithPublicUserIds []*struct {
		sschemas.Badge
		UserPublicId uuid.UUID `gorm:"column:user_public_id"`
	}
	result := db.Table(sschemas.Badge{}.TableName()+" b").
		Select("b.*, u.public_id as user_public_id").
		Joins(`LEFT JOIN "UsersToBadgesTable" utb ON utb.badge_id = b.id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = utb.user_id`).
		Where("u.public_id IN ?", uniquePublicIds).
		Find(&badgesWithPublicUserIds)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
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

	publicBadges := make([]*cgqlmodels.PublicBadge, len(publicIds))
	for _, badgeWithPublicUserId := range badgesWithPublicUserIds {
		for _, index := range publicIdToIndexesMap[badgeWithPublicUserId.UserPublicId] {
			publicBadges[index] = badgeWithPublicUserId.Badge.ToPublicBadge()
		}
	}

	return publicBadges, nil
}
