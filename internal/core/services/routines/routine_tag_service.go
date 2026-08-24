package routines

import (
	"context"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tags"
	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineTagServiceInterface interface {
	GetMyRoutineTagById(ctx context.Context, requestDto *apicontract.GetMyRoutineTagByIdRequestDto) (*apicontract.GetMyRoutineTagByIdResponseDto, *exceptions.Exception)
	GetAllMyRoutineTags(ctx context.Context, requestDto *apicontract.GetAllMyRoutineTagsRequestDto) (*apicontract.GetAllMyRoutineTagsResponseDto, *exceptions.Exception)
	CreateRoutineTag(ctx context.Context, requestDto *apicontract.CreateRoutineTagRequestDto) (*apicontract.CreateRoutineTagResponseDto, *exceptions.Exception)
	CreateRoutineTags(ctx context.Context, requestDto *apicontract.CreateRoutineTagsRequestDto) (*apicontract.CreateRoutineTagsResponseDto, *exceptions.Exception)
	UpdateMyRoutineTagById(ctx context.Context, requestDto *apicontract.UpdateMyRoutineTagByIdRequestDto) (*apicontract.UpdateMyRoutineTagByIdResponseDto, *exceptions.Exception)
	UpdateMyRoutineTagsByIds(ctx context.Context, requestDto *apicontract.UpdateMyRoutineTagsByIdsRequestDto) (*apicontract.UpdateMyRoutineTagsByIdsResponseDto, *exceptions.Exception)
	HardDeleteMyRoutineTagById(ctx context.Context, requestDto *apicontract.HardDeleteMyRoutineTagByIdRequestDto) (*apicontract.HardDeleteMyRoutineTagByIdResponseDto, *exceptions.Exception)
	HardDeleteMyRoutineTagsByIds(ctx context.Context, requestDto *apicontract.HardDeleteMyRoutineTagsByIdsRequestDto) (*apicontract.HardDeleteMyRoutineTagsByIdsResponseDto, *exceptions.Exception)

	SearchPrivateRoutineTags(ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchRoutineTagInput) (*gqlmodels.SearchRoutineTagConnection, *exceptions.Exception)
}

type RoutineTagService struct {
	validator            *validator.Validate
	db                   *gorm.DB
	routineTagRepository repositories.RoutineTagRepositoryInterface
}

func NewRoutineTagService(
	validator *validator.Validate,
	db *gorm.DB,
	routineTagRepository repositories.RoutineTagRepositoryInterface,
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

func convertRoutineTagIcon(icon *string) (*enums.SupportedIcon, *exceptions.Exception) {
	if icon == nil {
		return nil, nil
	}
	convertedIcon, err := enums.ConvertStringToSupportedIcon(*icon)
	if err != nil {
		return nil, exceptions.InvalidInput("RoutineTag").WithOrigin(err)
	}

	return convertedIcon, nil
}

func newRoutineTagResponseDto(routineTag schemas.RoutineTag) apicontract.RoutineTagResponseDto {
	var icon *string
	if routineTag.Icon != nil {
		iconValue := routineTag.Icon.String()
		icon = &iconValue
	}
	return apicontract.RoutineTagResponseDto{
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
	requestDto *apicontract.GetMyRoutineTagByIdRequestDto,
) (*apicontract.GetMyRoutineTagByIdResponseDto, *exceptions.Exception) {
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := newRoutineTagResponseDto(*routineTag)
	return &responseDto, nil
}

func (s *RoutineTagService) GetAllMyRoutineTags(
	ctx context.Context,
	requestDto *apicontract.GetAllMyRoutineTagsRequestDto,
) (*apicontract.GetAllMyRoutineTagsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}
	if requestDto.Param.AreDeleted != nil && *requestDto.Param.AreDeleted {
		responseDto := apicontract.GetAllMyRoutineTagsResponseDto{}
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(apicontract.GetAllMyRoutineTagsResponseDto, len(routineTags))
	for index, routineTag := range routineTags {
		responseDto[index] = newRoutineTagResponseDto(routineTag)
	}

	return &responseDto, nil
}

func (s *RoutineTagService) CreateRoutineTag(
	ctx context.Context,
	requestDto *apicontract.CreateRoutineTagRequestDto,
) (*apicontract.CreateRoutineTagResponseDto, *exceptions.Exception) {
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
		inputs.CreateRoutineTagInput{
			Id:    requestDto.Body.Id,
			Name:  requestDto.Body.Name,
			Color: requestDto.Body.Color,
			Icon:  icon,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.CreateRoutineTagResponseDto{
		Id:        *newRoutineTagId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) CreateRoutineTags(
	ctx context.Context,
	requestDto *apicontract.CreateRoutineTagsRequestDto,
) (*apicontract.CreateRoutineTagsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.CreateRoutineTagInput, len(requestDto.Body.CreatedRoutineTags))
	for index, createdRoutineTag := range requestDto.Body.CreatedRoutineTags {
		icon, exception := convertRoutineTagIcon(createdRoutineTag.Icon)
		if exception != nil {
			return nil, exception
		}
		input[index] = inputs.CreateRoutineTagInput{
			Id:    createdRoutineTag.Id,
			Name:  createdRoutineTag.Name,
			Color: createdRoutineTag.Color,
			Icon:  icon,
		}
	}
	newRoutineTagIds, exception := s.routineTagRepository.CreateMany(
		actorUserId,
		input,
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.CreateRoutineTagsResponseDto{
		Ids:       newRoutineTagIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) UpdateMyRoutineTagById(
	ctx context.Context,
	requestDto *apicontract.UpdateMyRoutineTagByIdRequestDto,
) (*apicontract.UpdateMyRoutineTagByIdResponseDto, *exceptions.Exception) {
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
		inputs.PartialUpdateRoutineTagInput{
			Values: inputs.UpdateRoutineTagInput{
				Name:  requestDto.Body.Values.Name,
				Color: requestDto.Body.Values.Color,
				Icon:  icon,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMyRoutineTagByIdResponseDto{
		UpdatedAt: updatedRoutineTag.UpdatedAt,
	}, nil
}

func (s *RoutineTagService) UpdateMyRoutineTagsByIds(
	ctx context.Context,
	requestDto *apicontract.UpdateMyRoutineTagsByIdsRequestDto,
) (*apicontract.UpdateMyRoutineTagsByIdsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewRoutineTagException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.UpdateRoutineTagByIdInput, len(requestDto.Body.UpdatedRoutineTags))
	for index, updatedRoutineTag := range requestDto.Body.UpdatedRoutineTags {
		icon, exception := convertRoutineTagIcon(updatedRoutineTag.Values.Icon)
		if exception != nil {
			return nil, exception
		}
		input[index] = inputs.UpdateRoutineTagByIdInput{
			Id: updatedRoutineTag.RoutineTagId,
			PartialUpdateInput: inputs.PartialUpdateInput[inputs.UpdateRoutineTagInput]{
				Values: inputs.UpdateRoutineTagInput{
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMyRoutineTagsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) HardDeleteMyRoutineTagById(
	ctx context.Context,
	requestDto *apicontract.HardDeleteMyRoutineTagByIdRequestDto,
) (*apicontract.HardDeleteMyRoutineTagByIdResponseDto, *exceptions.Exception) {
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.HardDeleteMyRoutineTagByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineTagService) HardDeleteMyRoutineTagsByIds(
	ctx context.Context,
	requestDto *apicontract.HardDeleteMyRoutineTagsByIdsRequestDto,
) (*apicontract.HardDeleteMyRoutineTagsByIdsResponseDto, *exceptions.Exception) {
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.HardDeleteMyRoutineTagsByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for GraphQL RoutineTag ============================== */

func (s *RoutineTagService) SearchPrivateRoutineTags(
	ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchRoutineTagInput,
) (*gqlmodels.SearchRoutineTagConnection, *exceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	query := db.Model(&schemas.RoutineTag{}).
		Select(`"RoutineTagTable".*`).
		Where(`"RoutineTagTable".owner_id = ?`, userId)

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := searchcursor.Decode[gqlmodels.SearchRoutineTagCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = gqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == gqlmodels.SearchSortOrderDesc {
			cending = gqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case gqlmodels.SearchRoutineTagSortByName:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchRoutineTagSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchRoutineTagSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := constants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, constants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var routineTags []schemas.RoutineTag
	if err := query.Find(&routineTags).Error; err != nil {
		return nil, apiexceptions.NewRoutineTagException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(routineTags) > limit
	searchEdges := make([]*gqlmodels.SearchRoutineTagEdge, len(routineTags))

	for index, routineTag := range routineTags {
		searchCursor := searchcursor.SearchCursor[gqlmodels.SearchRoutineTagCursorFields]{
			Fields: gqlmodels.SearchRoutineTagCursorFields{
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

		searchEdges[index] = &gqlmodels.SearchRoutineTagEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routineTag.ToPrivateRoutineTag(),
		}
	}

	searchPageInfo := &gqlmodels.SearchPageInfo{
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

	return &gqlmodels.SearchRoutineTagConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
