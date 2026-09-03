package repositories

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskDependencyRepositoryInterface interface {
	GetAllByRoutineId(
		routineId uuid.UUID,
		opts ...RepositoryOptions,
	) ([]schemas.RoutineTaskDependency, *cexceptions.Exception)
	CreateOneByRoutineId(
		routineId uuid.UUID,
		input inputs.CreateRoutineTaskDependencyInput,
		opts ...RepositoryOptions,
	) (*schemas.RoutineTaskDependency, *cexceptions.Exception)
	CreateManyByRoutineId(
		routineId uuid.UUID,
		input []inputs.CreateRoutineTaskDependencyInput,
		opts ...RepositoryOptions,
	) *cexceptions.Exception
	UpdateOneByRoutineId(
		routineId uuid.UUID,
		input inputs.UpdateRoutineTaskDependencyInput,
		opts ...RepositoryOptions,
	) *cexceptions.Exception
	UpdateManyByRoutineId(
		routineId uuid.UUID,
		input []inputs.UpdateRoutineTaskDependencyInput,
		opts ...RepositoryOptions,
	) *cexceptions.Exception
	DeleteOneByRoutineId(
		routineId uuid.UUID,
		input inputs.RoutineTaskDependencyKey,
		opts ...RepositoryOptions,
	) (int64, *cexceptions.Exception)
	DeleteManyByRoutineId(
		routineId uuid.UUID,
		input []inputs.RoutineTaskDependencyKey,
		opts ...RepositoryOptions,
	) (int64, *cexceptions.Exception)
}

type RoutineTaskDependencyRepository struct {
	db         *gorm.DB
	exceptions exceptions.RoutineTaskDependencyException
}

func NewRoutineTaskDependencyRepository(db *gorm.DB) RoutineTaskDependencyRepositoryInterface {
	return &RoutineTaskDependencyRepository{
		db:         db,
		exceptions: exceptions.NewRoutineTaskDependencyException(),
	}
}

func (r *RoutineTaskDependencyRepository) incrementRoutineDefinitionVersion(
	routineId uuid.UUID,
	db *gorm.DB,
) *cexceptions.Exception {
	result := db.
		Model(&schemas.Routine{}).
		Where("id = ?", routineId).
		UpdateColumn("definition_version", gorm.Expr("definition_version + ?", 1))
	if result.Error != nil {
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}
	result = db.
		Model(&schemas.Routine{}).
		Where("id = ? AND status IN ?", routineId, []cenums.RoutineStatus{
			cenums.RoutineStatus_Completed,
			cenums.RoutineStatus_OverDue,
		}).
		Update("status", cenums.RoutineStatus_Scheduled)
	if result.Error != nil {
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}
	return nil
}

func (r *RoutineTaskDependencyRepository) validateRoutineTaskDependencyInput(
	routineId uuid.UUID,
	routineTaskId uuid.UUID,
	previousRoutineTaskId uuid.UUID,
) *cexceptions.Exception {
	if routineId == uuid.Nil || routineTaskId == uuid.Nil || previousRoutineTaskId == uuid.Nil {
		return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("routine and routine task ids are required"))
	}
	if routineTaskId == previousRoutineTaskId {
		return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("a routine task cannot depend on itself"))
	}
	return nil
}

func (r *RoutineTaskDependencyRepository) GetAllByRoutineId(
	routineId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.RoutineTaskDependency, *cexceptions.Exception) {
	if routineId == uuid.Nil {
		return nil, r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("routine id is required"))
	}

	parsedOptions := ParseRepositoryOptions(append([]RepositoryOptions{WithDB(r.db)}, opts...)...)
	var dependencies []schemas.RoutineTaskDependency
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskDependency{}).
		Joins(`INNER JOIN "RoutineTaskTable" task ON task.id = "RoutineDependencyTable".routine_task_id`).
		Where(`task.routine_id = ?`, routineId).
		Order(`"RoutineDependencyTable".created_at ASC`).
		Find(&dependencies)
	if result.Error != nil {
		return nil, r.exceptions.FailedToGet().WithOrigin(result.Error)
	}

	return dependencies, nil
}

func (r *RoutineTaskDependencyRepository) CreateOneByRoutineId(
	routineId uuid.UUID,
	input inputs.CreateRoutineTaskDependencyInput,
	opts ...RepositoryOptions,
) (*schemas.RoutineTaskDependency, *cexceptions.Exception) {
	if exception := r.validateRoutineTaskDependencyInput(routineId, input.RoutineTaskId, input.PreviousRoutineTaskId); exception != nil {
		return nil, exception
	}

	parsedOptions := ParseRepositoryOptions(append([]RepositoryOptions{WithDB(r.db)}, opts...)...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	dependency := schemas.RoutineTaskDependency{
		RoutineTaskId:         input.RoutineTaskId,
		PreviousRoutineTaskId: input.PreviousRoutineTaskId,
		Description:           input.Description,
		Progress:              input.Progress,
	}
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskDependency{}).
		Create(&dependency)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") ||
			strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return nil, r.exceptions.DependencyAlreadyExists().WithOrigin(result.Error)
		}
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	if exception := r.incrementRoutineDefinitionVersion(routineId, parsedOptions.DB); exception != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, exception
	}
	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &dependency, nil
}

func (r *RoutineTaskDependencyRepository) CreateManyByRoutineId(
	routineId uuid.UUID,
	input []inputs.CreateRoutineTaskDependencyInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return r.exceptions.NoChanges()
	}
	for _, dependency := range input {
		if exception := r.validateRoutineTaskDependencyInput(
			routineId,
			dependency.RoutineTaskId,
			dependency.PreviousRoutineTaskId,
		); exception != nil {
			return exception
		}
	}

	parsedOptions := ParseRepositoryOptions(append([]RepositoryOptions{WithDB(r.db)}, opts...)...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	dependencies := make([]schemas.RoutineTaskDependency, len(input))
	for index, dependency := range input {
		dependencies[index] = schemas.RoutineTaskDependency{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
			Description:           dependency.Description,
			Progress:              dependency.Progress,
		}
	}
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskDependency{}).
		CreateInBatches(&dependencies, parsedOptions.BatchSize)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") ||
			strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return r.exceptions.DependencyAlreadyExists().WithOrigin(result.Error)
		}
		return r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}
	if exception := r.incrementRoutineDefinitionVersion(routineId, parsedOptions.DB); exception != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return exception
	}
	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *RoutineTaskDependencyRepository) UpdateOneByRoutineId(
	routineId uuid.UUID,
	input inputs.UpdateRoutineTaskDependencyInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	return r.UpdateManyByRoutineId(routineId, []inputs.UpdateRoutineTaskDependencyInput{input}, opts...)
}

func (r *RoutineTaskDependencyRepository) UpdateManyByRoutineId(
	routineId uuid.UUID,
	input []inputs.UpdateRoutineTaskDependencyInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(append([]RepositoryOptions{WithDB(r.db)}, opts...)...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	dependencies := make([]schemas.RoutineTaskDependency, len(input))
	for index, dependency := range input {
		if exception := r.validateRoutineTaskDependencyInput(
			routineId,
			dependency.RoutineTaskId,
			dependency.PreviousRoutineTaskId,
		); exception != nil {
			if shouldStartTransaction {
				parsedOptions.DB.Rollback()
			}
			return exception
		}
		dependencies[index] = schemas.RoutineTaskDependency{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
			Description:           dependency.Description,
			Progress:              dependency.Progress,
		}
	}
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskDependency{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "routine_task_id",
				},
				{
					Name: "previous_routine_task_id",
				},
			},
			DoUpdates: clause.AssignmentColumns([]string{"description", "progress", "updated_at"}),
		}).
		CreateInBatches(&dependencies, parsedOptions.BatchSize)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}
	if exception := r.incrementRoutineDefinitionVersion(routineId, parsedOptions.DB); exception != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return exception
	}
	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *RoutineTaskDependencyRepository) DeleteOneByRoutineId(
	routineId uuid.UUID,
	input inputs.RoutineTaskDependencyKey,
	opts ...RepositoryOptions,
) (int64, *cexceptions.Exception) {
	return r.DeleteManyByRoutineId(routineId, []inputs.RoutineTaskDependencyKey{input}, opts...)
}

func (r *RoutineTaskDependencyRepository) DeleteManyByRoutineId(
	routineId uuid.UUID,
	input []inputs.RoutineTaskDependencyKey,
	opts ...RepositoryOptions,
) (int64, *cexceptions.Exception) {
	if len(input) == 0 {
		return 0, r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(append([]RepositoryOptions{WithDB(r.db)}, opts...)...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	query := parsedOptions.DB.Model(&schemas.RoutineTaskDependency{})
	for index, dependency := range input {
		condition := "routine_task_id = ? AND previous_routine_task_id = ?"
		if index == 0 {
			query = query.Where(condition, dependency.RoutineTaskId, dependency.PreviousRoutineTaskId)
			continue
		}
		query = query.Or(condition, dependency.RoutineTaskId, dependency.PreviousRoutineTaskId)
	}
	result := query.Delete(&schemas.RoutineTaskDependency{})
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return 0, r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected > 0 {
		if exception := r.incrementRoutineDefinitionVersion(routineId, parsedOptions.DB); exception != nil {
			if shouldStartTransaction {
				parsedOptions.DB.Rollback()
			}
			return 0, exception
		}
	}
	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return 0, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return result.RowsAffected, nil
}
