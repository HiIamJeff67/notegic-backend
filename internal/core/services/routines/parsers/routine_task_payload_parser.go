package parsers

import (
	"fmt"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	concurrency "github.com/HiIamJeff67/notegic-backend/shared/lib/concurrency"
	jsonpayload "github.com/HiIamJeff67/notegic-backend/shared/lib/jsonpayload"

	editableblock "github.com/HiIamJeff67/notegic-backend/shared/util/editableblock"

	routinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	blocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type RoutineTaskPayloadParserInterface interface {
	ValidateRoutineTaskPayload(
		purpose enums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *exceptions.Exception
}

type RoutineTaskPayloadParser struct {
	validator *validator.Validate
}

func NewRoutineTaskPayloadParser(validatorInstance *validator.Validate) RoutineTaskPayloadParserInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	return &RoutineTaskPayloadParser{validator: validatorInstance}
}

func DecodePayload[T any](validatorInstance *validator.Validate, task schemas.RoutineTask) (*T, *exceptions.Exception) {
	var payload T
	if err := jsonpayload.Decode(task.Payload, &payload); err != nil {
		return nil, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if err := validatorInstance.Struct(payload); err != nil {
		return nil, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	return &payload, nil
}

func FlattenArborizedBlock(
	blockPackId uuid.UUID,
	arborizedEditableBlock *blocknote.ArborizedEditableBlock,
) ([]schemas.Block, []uuid.UUID, int64, *exceptions.Exception) {
	if blockPackId == uuid.Nil {
		return nil, nil, 0, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(fmt.Errorf("blockPackId is required"))
	}
	rawFlattenedBlocks, totalSize, err := editableblock.FlattenEditableBlock(arborizedEditableBlock)
	if err != nil {
		return nil, nil, 0, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if len(rawFlattenedBlocks) == 0 {
		return nil, nil, 0, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(fmt.Errorf("arborizedEditableBlock must contain at least one block"))
	}

	blocks := make([]schemas.Block, len(rawFlattenedBlocks))
	blockIds := make([]uuid.UUID, len(rawFlattenedBlocks))
	for index, rawFlattenedBlock := range rawFlattenedBlocks {
		blockType := enums.BlockType(rawFlattenedBlock.Type)
		if rawFlattenedBlock.Id == uuid.Nil || !blockType.IsValidEnum() {
			return nil, nil, 0, exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Resolve",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(fmt.Errorf("invalid arborizedEditableBlock at flattened index %d", index))
		}

		blockIds[index] = rawFlattenedBlock.Id
		blocks[index] = schemas.Block{
			Id:            rawFlattenedBlock.Id,
			BlockPackId:   blockPackId,
			ParentBlockId: rawFlattenedBlock.ParentBlockId,
			PrevBlockId:   rawFlattenedBlock.PrevBlockId,
			NextBlockId:   rawFlattenedBlock.NextBlockId,
			Type:          enums.BlockType(rawFlattenedBlock.Type),
			Props:         datatypes.JSON(rawFlattenedBlock.Props),
			Content:       datatypes.JSON(rawFlattenedBlock.Content),
		}
	}
	return blocks, blockIds, totalSize, nil
}

func (s *RoutineTaskPayloadParser) ValidateRoutineTaskPayload(
	purpose enums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *exceptions.Exception {
	switch purpose {
	case enums.RoutineTaskPurpose_CreateRootShelf:
		var parsedPayload routinetasktypes.CreateRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateRootShelf:
		var parsedPayload routinetasktypes.UpdateRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_ResetRootShelf:
		var parsedPayload routinetasktypes.ResetRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateSubShelf:
		var parsedPayload routinetasktypes.CreateSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateSubShelf:
		var parsedPayload routinetasktypes.UpdateSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_ResetSubShelf:
		var parsedPayload routinetasktypes.ResetSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateBlockPack:
		var parsedPayload routinetasktypes.CreateBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}

		validateBlockDto := make([]blocknote.ArborizedEditableBlock, len(parsedPayload.Template.Blocks))
		for index, block := range parsedPayload.Template.Blocks {
			validateBlockDto[index] = block.ArborizedEditableBlock
		}

		validateBlockFunc := func(validateDto blocknote.ArborizedEditableBlock) (bool, error) {
			if exception := validateArborizedEditableBlock(&validateDto); exception != nil {
				return false, exception
			}
			return true, nil
		}

		validateBlockResults := concurrency.Execute(
			validateBlockDto,
			min(10, max(len(validateBlockDto)/10, len(validateBlockDto)%10)),
			validateBlockFunc,
		)

		for _, validateBlockResult := range validateBlockResults {
			if validateBlockResult.Err != nil {
				return exceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).
					WithOrigin(fmt.Errorf(
						"invalid template.blocks[%d].arborizedEditableBlock: %w",
						validateBlockResult.Index,
						validateBlockResult.Err,
					))
			}
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateBlockPack:
		var parsedPayload routinetasktypes.UpdateBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}

		for index, updatedBlock := range parsedPayload.UpdatedBlocks {
			if exception := validateArborizedEditableBlock(updatedBlock.ArborizedEditableBlock); exception != nil {
				return exceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).
					WithOrigin(fmt.Errorf(
						"invalid updatedBlocks[%d].arborizedEditableBlock: %w",
						index,
						exception,
					))
			}
			if len(updatedBlock.ArborizedEditableBlock.Children) > 0 {
				return exceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).
					WithOrigin(fmt.Errorf(
						"invalid updatedBlocks[%d].arborizedEditableBlock: children are not allowed for update operations",
						index,
					))
			}
		}
		return nil

	case enums.RoutineTaskPurpose_ResetBlockPack:
		var parsedPayload routinetasktypes.ResetBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateRoutine:
		var parsedPayload routinetasktypes.CreateRoutineRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateRoutine:
		var parsedPayload routinetasktypes.UpdateRoutineRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return exceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	default:
		return exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("unsupported routine task purpose: %s", purpose))
	}
}

func validateArborizedEditableBlock(
	arborizedEditableBlock *blocknote.ArborizedEditableBlock,
) *exceptions.Exception {
	if arborizedEditableBlock == nil {
		return exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(fmt.Errorf("arborizedEditableBlock is required"))
	}

	rawFlattenedBlocks, _, err := editableblock.FlattenEditableBlock(arborizedEditableBlock)
	if err != nil {
		return exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if len(rawFlattenedBlocks) == 0 {
		return exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("arborizedEditableBlock must contain at least one block"))
	}

	return nil
}
