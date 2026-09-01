package parsers

import (
	"fmt"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconcurrency "github.com/HiIamJeff67/notegic-backend/shared/lib/concurrency"
	sjsonpayload "github.com/HiIamJeff67/notegic-backend/shared/lib/jsonpayload"
	seditableblock "github.com/HiIamJeff67/notegic-backend/shared/util/editableblock"
)

type RoutineTaskPayloadParserInterface interface {
	ValidateRoutineTaskPayload(
		purpose cenums.RoutineTaskPurpose,
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

	rawFlattenedBlocks, _, err := seditableblock.FlattenEditableBlock(arborizedEditableBlock)
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

func (s *RoutineTaskPayloadParser) ValidateRoutineTaskPayload(
	purpose cenums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *cexceptions.Exception {
	switch purpose {
	case cenums.RoutineTaskPurpose_GetSubShelf:
		var parsedPayload croutinetasktypes.GetSubShelfRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_DeleteSubShelf:
		var parsedPayload croutinetasktypes.DeleteSubShelfRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_GetBlockPack:
		var parsedPayload croutinetasktypes.GetBlockPackRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_DeleteBlockPack:
		var parsedPayload croutinetasktypes.DeleteBlockPackRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_GetRoutine:
		var parsedPayload croutinetasktypes.GetRoutineRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_DeleteRoutine:
		var parsedPayload croutinetasktypes.DeleteRoutineRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_GetMaterial:
		var parsedPayload croutinetasktypes.GetMaterialRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
	case cenums.RoutineTaskPurpose_CreateMaterial:
		var parsedPayload croutinetasktypes.CreateMaterialRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
		if err := parsedPayload.ParentSubShelfId.Validate(); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		return nil
	case cenums.RoutineTaskPurpose_UpdateMaterial:
		var parsedPayload croutinetasktypes.UpdateMaterialRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
		if parsedPayload.Name == nil && parsedPayload.Size == nil &&
			parsedPayload.ContentKey == nil && parsedPayload.ContentType == nil &&
			parsedPayload.ParseMediaType == nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(fmt.Errorf("at least one material field must be provided for update"))
		}
		return nil
	case cenums.RoutineTaskPurpose_DeleteMaterial:
		var parsedPayload croutinetasktypes.DeleteMaterialRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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

	case cenums.RoutineTaskPurpose_CreateSubShelf:
		var parsedPayload croutinetasktypes.CreateSubShelfRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
		if err := parsedPayload.FakeId.Validate(); err != nil {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		if parsedPayload.PrevSubShelfId != nil {
			if err := parsedPayload.PrevSubShelfId.Validate(); err != nil {
				return cexceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).WithOrigin(err)
			}
		}
		return nil

	case cenums.RoutineTaskPurpose_UpdateSubShelf:
		var parsedPayload croutinetasktypes.UpdateSubShelfRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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

	case cenums.RoutineTaskPurpose_CreateBlockPack:
		var parsedPayload croutinetasktypes.CreateBlockPackRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
		if err := parsedPayload.TargetSubShelfId.Validate(); err != nil {
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

		validateBlockResults := sconcurrency.Execute(
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

	case cenums.RoutineTaskPurpose_UpdateBlockPack:
		var parsedPayload croutinetasktypes.UpdateBlockPackRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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

		if len(parsedPayload.Blocks) > croutinetasktypes.MaxRoutineTaskBlockPackUpdates {
			return cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Parse",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(fmt.Errorf(
				"blocks must contain at most %d blocks",
				croutinetasktypes.MaxRoutineTaskBlockPackUpdates,
			))
		}
		seenBlockIds := make(map[uuid.UUID]struct{}, len(parsedPayload.Blocks))
		for index, updatedBlock := range parsedPayload.Blocks {
			if _, exists := seenBlockIds[updatedBlock.BlockId]; exists {
				return cexceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).WithOrigin(fmt.Errorf("blocks[%d].blockId is duplicated", index))
			}
			seenBlockIds[updatedBlock.BlockId] = struct{}{}
			if exception := validateArborizedEditableBlock(updatedBlock.ArborizedEditableBlock); exception != nil {
				return cexceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).
					WithOrigin(fmt.Errorf(
						"invalid blocks[%d].arborizedEditableBlock: %w",
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
						"invalid blocks[%d].arborizedEditableBlock: children are not allowed for update operations",
						index,
					))
			}
			if updatedBlock.ArborizedEditableBlock.Id != updatedBlock.BlockId {
				return cexceptions.New(
					"InvalidRoutineTaskPayload",
					"RoutineTask",
					"Parse",
					"Routine task payload is invalid",
					http.StatusBadRequest,
				).WithOrigin(fmt.Errorf(
					"invalid blocks[%d].arborizedEditableBlock: id must match blockId",
					index,
				))
			}
		}
		return nil

	case cenums.RoutineTaskPurpose_CreateRoutine:
		var parsedPayload croutinetasktypes.CreateRoutineRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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

	case cenums.RoutineTaskPurpose_UpdateRoutine:
		var parsedPayload croutinetasktypes.UpdateRoutineRoutineTaskPayload
		if err := sjsonpayload.Decode(payload, &parsedPayload); err != nil {
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
