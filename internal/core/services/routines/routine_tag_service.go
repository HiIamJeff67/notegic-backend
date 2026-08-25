package routines

import (
	"context"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tags"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineTagServiceInterface interface {
	GetMyRoutineTagById(ctx context.Context, requestDto *capi.GetMyRoutineTagByIdRequestDto) (*capi.GetMyRoutineTagByIdResponseDto, *cexceptions.Exception)
	GetAllMyRoutineTags(ctx context.Context, requestDto *capi.GetAllMyRoutineTagsRequestDto) (*capi.GetAllMyRoutineTagsResponseDto, *cexceptions.Exception)
	CreateRoutineTag(ctx context.Context, requestDto *capi.CreateRoutineTagRequestDto) (*capi.CreateRoutineTagResponseDto, *cexceptions.Exception)
	CreateRoutineTags(ctx context.Context, requestDto *capi.CreateRoutineTagsRequestDto) (*capi.CreateRoutineTagsResponseDto, *cexceptions.Exception)
	UpdateMyRoutineTagById(ctx context.Context, requestDto *capi.UpdateMyRoutineTagByIdRequestDto) (*capi.UpdateMyRoutineTagByIdResponseDto, *cexceptions.Exception)
	UpdateMyRoutineTagsByIds(ctx context.Context, requestDto *capi.UpdateMyRoutineTagsByIdsRequestDto) (*capi.UpdateMyRoutineTagsByIdsResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTagById(ctx context.Context, requestDto *capi.HardDeleteMyRoutineTagByIdRequestDto) (*capi.HardDeleteMyRoutineTagByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTagsByIds(ctx context.Context, requestDto *capi.HardDeleteMyRoutineTagsByIdsRequestDto) (*capi.HardDeleteMyRoutineTagsByIdsResponseDto, *cexceptions.Exception)

	SearchPrivateRoutineTags(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTagInput) (*cgqlmodels.SearchRoutineTagConnection, *cexceptions.Exception)
}

type RoutineTagService struct {
	validator            *validator.Validate
	db                   *gorm.DB
	routineTagRepository srepositories.RoutineTagRepositoryInterface
}

func NewRoutineTagService(
	validator *validator.Validate,
	db *gorm.DB,
	routineTagRepository srepositories.RoutineTagRepositoryInterface,
) RoutineTagServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &RoutineTagService{
		validator:            validator,
		db:                   db,
		routineTagRepository: routineTagRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func convertRoutineTagIcon(icon *string) (*cenums.SupportedIcon, *cexceptions.Exception) {
	if icon == nil {
		return nil, nil
	}
	convertedIcon, err := cenums.ConvertStringToSupportedIcon(*icon)
	if err != nil {
		return nil, cexceptions.InvalidInput("RoutineTag").WithOrigin(err)
	}

	return convertedIcon, nil
}

func newRoutineTagResponseDto(routineTag sschemas.RoutineTag) capi.RoutineTagResponseDto {
	var icon *string
	if routineTag.Icon != nil {
		iconValue := routineTag.Icon.String()
		icon = &iconValue
	}
	return capi.RoutineTagResponseDto{
		Id:        routineTag.Id,
		Name:      routineTag.Name,
		Color:     routineTag.Color,
		Icon:      icon,
		UpdatedAt: routineTag.UpdatedAt,
		CreatedAt: routineTag.CreatedAt,
	}
}

/* ============================== Service Methods for RoutineTag ============================== */

func (s *RoutineTagService) GetMyRoutineTagById(
	ctx context.Context,
	requestDto *capi.GetMyRoutineTagByIdRequestDto,
) (*capi.GetMyRoutineTagByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}
	if requestDto.Param.IsDeleted != nil && *requestDto.Param.IsDeleted {
		return nil, apiexceptions.NewRoutineTagException().NotFound()
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTag, exception := s.routineTagRepository.GetOneById(
		requestDto.Param.RoutineTagId,
		actorUserId,
		nil,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := newRoutineTagResponseDto(*routineTag)
	return &responseDto, nil
}

func (s *RoutineTagService) GetAllMyRoutineTags(
	ctx context.Context,
	requestDto *capi.GetAllMyRoutineTagsRequestDto,
) (*capi.GetAllMyRoutineTagsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}
	if requestDto.Param.AreDeleted != nil && *requestDto.Param.AreDeleted {
		responseDto := capi.GetAllMyRoutineTagsResponseDto{}
		return &responseDto, nil
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTags, exception := s.routineTagRepository.GetAllByUserId(
		actorUserId,
		nil,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(capi.GetAllMyRoutineTagsResponseDto, len(routineTags))
	for index, routineTag := range routineTags {
		responseDto[index] = newRoutineTagResponseDto(routineTag)
	}

	return &responseDto, nil
}

func (s *RoutineTagService) CreateRoutineTag(
	ctx context.Context,
	requestDto *capi.CreateRoutineTagRequestDto,
) (*capi.CreateRoutineTagResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	icon, exception := convertRoutineTagIcon(requestDto.Body.Icon)
	if exception != nil {
		return nil, exception
	}

	newRoutineTagId, exception := s.routineTagRepository.CreateOne(
		actorUserId,
		sinputs.CreateRoutineTagInput{
			Id:    requestDto.Body.Id,
			Name:  requestDto.Body.Name,
			Color: requestDto.Body.Color,
			Icon:  icon,
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRoutineTagResponseDto{
		Id:        *newRoutineTagId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) CreateRoutineTags(
	ctx context.Context,
	requestDto *capi.CreateRoutineTagsRequestDto,
) (*capi.CreateRoutineTagsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.CreateRoutineTagInput, len(requestDto.Body.CreatedRoutineTags))
	for index, createdRoutineTag := range requestDto.Body.CreatedRoutineTags {
		icon, exception := convertRoutineTagIcon(createdRoutineTag.Icon)
		if exception != nil {
			return nil, exception
		}
		input[index] = sinputs.CreateRoutineTagInput{
			Id:    createdRoutineTag.Id,
			Name:  createdRoutineTag.Name,
			Color: createdRoutineTag.Color,
			Icon:  icon,
		}
	}
	newRoutineTagIds, exception := s.routineTagRepository.CreateMany(
		actorUserId,
		input,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRoutineTagsResponseDto{
		Ids:       newRoutineTagIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) UpdateMyRoutineTagById(
	ctx context.Context,
	requestDto *capi.UpdateMyRoutineTagByIdRequestDto,
) (*capi.UpdateMyRoutineTagByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	icon, exception := convertRoutineTagIcon(requestDto.Body.Values.Icon)
	if exception != nil {
		return nil, exception
	}

	updatedRoutineTag, exception := s.routineTagRepository.UpdateOneById(
		requestDto.Param.RoutineTagId,
		actorUserId,
		sinputs.PartialUpdateRoutineTagInput{
			Values: sinputs.UpdateRoutineTagInput{
				Name:  requestDto.Body.Values.Name,
				Color: requestDto.Body.Values.Color,
				Icon:  icon,
			},
			SetNull: requestDto.Body.SetNull,
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRoutineTagByIdResponseDto{
		UpdatedAt: updatedRoutineTag.UpdatedAt,
	}, nil
}

func (s *RoutineTagService) UpdateMyRoutineTagsByIds(
	ctx context.Context,
	requestDto *capi.UpdateMyRoutineTagsByIdsRequestDto,
) (*capi.UpdateMyRoutineTagsByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.UpdateRoutineTagByIdInput, len(requestDto.Body.UpdatedRoutineTags))
	for index, updatedRoutineTag := range requestDto.Body.UpdatedRoutineTags {
		icon, exception := convertRoutineTagIcon(updatedRoutineTag.Values.Icon)
		if exception != nil {
			return nil, exception
		}
		input[index] = sinputs.UpdateRoutineTagByIdInput{
			Id: updatedRoutineTag.RoutineTagId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateRoutineTagInput]{
				Values: sinputs.UpdateRoutineTagInput{
					Name:  updatedRoutineTag.Values.Name,
					Color: updatedRoutineTag.Values.Color,
					Icon:  icon,
				},
				SetNull: updatedRoutineTag.SetNull,
			},
		}
	}
	exception = s.routineTagRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRoutineTagsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) HardDeleteMyRoutineTagById(
	ctx context.Context,
	requestDto *capi.HardDeleteMyRoutineTagByIdRequestDto,
) (*capi.HardDeleteMyRoutineTagByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineTagRepository.HardDeleteOneById(
		requestDto.Param.RoutineTagId,
		actorUserId,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutineTagByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) HardDeleteMyRoutineTagsByIds(
	ctx context.Context,
	requestDto *capi.HardDeleteMyRoutineTagsByIdsRequestDto,
) (*capi.HardDeleteMyRoutineTagsByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineTagRepository.HardDeleteManyByIds(
		requestDto.Body.RoutineTagIds,
		actorUserId,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutineTagsByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for GraphQL RoutineTag ============================== */

func (s *RoutineTagService) SearchPrivateRoutineTags(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTagInput,
) (*cgqlmodels.SearchRoutineTagConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	query := db.Model(&sschemas.RoutineTag{}).
		Select(`"RoutineTagTable".*`).
		Where(`"RoutineTagTable".owner_id = ?`, userId)

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRoutineTagCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRoutineTagSortByName:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTagSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTagSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var routineTags []sschemas.RoutineTag
	if err := query.Find(&routineTags).Error; err != nil {
		return nil, apiexceptions.NewRoutineTagException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(routineTags) > limit
	searchEdges := make([]*cgqlmodels.SearchRoutineTagEdge, len(routineTags))

	for index, routineTag := range routineTags {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRoutineTagCursorFields]{
			Fields: cgqlmodels.SearchRoutineTagCursorFields{
				ID: routineTag.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchRoutineTagEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routineTag.ToPrivateRoutineTag(),
		}
	}

	searchPageInfo := &cgqlmodels.SearchPageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0,
	}

	if len(searchEdges) > 0 {
		searchPageInfo.StartEncodedSearchCursor = &searchEdges[0].EncodedSearchCursor
		searchPageInfo.EndEncodedSearchCursor = &searchEdges[len(searchEdges)-1].EncodedSearchCursor
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	if hasNextPage {
		searchEdges = searchEdges[:limit]
	}

	return &cgqlmodels.SearchRoutineTagConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
