package parsers

import (
	"fmt"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	concurrency "github.com/HiIamJeff67/notegic-backend/shared/lib/concurrency"
	jsonpayload "github.com/HiIamJeff67/notegic-backend/shared/lib/jsonpayload"

	editableblock "github.com/HiIamJeff67/notegic-backend/shared/util/editableblock"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskPayloadParserInterface interface {
	ValidateRoutineTaskPayload(
		purpose enums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *cexceptions.Exception
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

func DecodePayload[T any](validatorInstance *validator.Validate, task schemas.RoutineTask) (*T, *cexceptions.Exception) {
	var payload T
	if err := jsonpayload.Decode(task.Payload, &payload); err != nil {
		return nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if err := validatorInstance.Struct(payload); err != nil {
		return nil, cexceptions.New(
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
	arborizedEditableBlock *cblocknote.ArborizedEditableBlock,
) ([]schemas.Block, []uuid.UUID, int64, *cexceptions.Exception) {
	if blockPackId == uuid.Nil {
		return nil, nil, 0, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(fmt.Errorf("blockPackId is required"))
	}
	rawFlattenedBlocks, totalSize, err := editableblock.FlattenEditableBlock(arborizedEditableBlock)
	if err != nil {
		return nil, nil, 0, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if len(rawFlattenedBlocks) == 0 {
		return nil, nil, 0, cexceptions.New(
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
			return nil, nil, 0, cexceptions.New(
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
) *cexceptions.Exception {
	switch purpose {
	case enums.RoutineTaskPurpose_CreateRootShelf:
		var parsedPayload croutinetasktypes.CreateRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateRootShelf:
		var parsedPayload croutinetasktypes.UpdateRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_ResetRootShelf:
		var parsedPayload croutinetasktypes.ResetRootShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateSubShelf:
		var parsedPayload croutinetasktypes.CreateSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateSubShelf:
		var parsedPayload croutinetasktypes.UpdateSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_ResetSubShelf:
		var parsedPayload croutinetasktypes.ResetSubShelfRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateBlockPack:
		var parsedPayload croutinetasktypes.CreateBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}

		validateBlockDto := make([]cblocknote.ArborizedEditableBlock, len(parsedPayload.Template.Blocks))
		for index, block := range parsedPayload.Template.Blocks {
			validateBlockDto[index] = block.ArborizedEditableBlock
		}

		validateBlockFunc := func(validateDto cblocknote.ArborizedEditableBlock) (bool, error) {
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
				return cexceptions.New(
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
		var parsedPayload croutinetasktypes.UpdateBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}

		for index, updatedBlock := range parsedPayload.UpdatedBlocks {
			if exception := validateArborizedEditableBlock(updatedBlock.ArborizedEditableBlock); exception != nil {
				return cexceptions.New(
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
				return cexceptions.New(
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
		var parsedPayload croutinetasktypes.ResetBlockPackRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_CreateRoutine:
		var parsedPayload croutinetasktypes.CreateRoutineRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	case enums.RoutineTaskPurpose_UpdateRoutine:
		var parsedPayload croutinetasktypes.UpdateRoutineRoutineTaskPayload
		if err := jsonpayload.Decode(payload, &parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if err := s.validator.Struct(&parsedPayload); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil

	default:
		return cexceptions.New(
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
	arborizedEditableBlock *cblocknote.ArborizedEditableBlock,
) *cexceptions.Exception {
	if arborizedEditableBlock == nil {
		return cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(fmt.Errorf("arborizedEditableBlock is required"))
	}

	rawFlattenedBlocks, _, err := editableblock.FlattenEditableBlock(arborizedEditableBlock)
	if err != nil {
		return cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Parse",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	if len(rawFlattenedBlocks) == 0 {
		return cexceptions.New(
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
