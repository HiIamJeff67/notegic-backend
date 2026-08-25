package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type InboxEventRepositoryInterface interface {
	CreateOne(input inputs.CreateInboxEventInput, opts RepositoryOptionFields) (bool, *cexceptions.Exception)
}

type InboxEventRepository struct {
	exceptions exceptions.InboxEventException
}

func NewInboxEventRepository(repositoryExceptions ...exceptions.InboxEventException) InboxEventRepositoryInterface {
	repositoryException := exceptions.NewInboxEventException()
	if len(repositoryExceptions) > 0 {
		repositoryException = repositoryExceptions[0]
	}

	return &InboxEventRepository{exceptions: repositoryException}
}

func (r *InboxEventRepository) CreateOne(
	input inputs.CreateInboxEventInput,
	parsedOptions RepositoryOptionFields,
) (bool, *cexceptions.Exception) {
	if input.EventId == uuid.Nil {
		return false, r.exceptions.EventIdRequired()
	}
	if parsedOptions.DB == nil {
		return false, r.exceptions.DatabaseUnavailable()
	}

	result := parsedOptions.DB.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&schemas.InboxEvent{EventId: input.EventId})
	if result.Error != nil {
		return false, r.exceptions.FailedToRecord().WithOrigin(result.Error)
	}

	return result.RowsAffected > 0, nil
}
