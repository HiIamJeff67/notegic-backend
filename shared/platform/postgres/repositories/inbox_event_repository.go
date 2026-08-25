package repositories

import (
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type InboxEventRepositoryInterface interface {
	CreateOne(input inputs.CreateInboxEventInput, opts RepositoryOptionFields) (bool, *cexceptions.Exception)
}

type InboxEventRepository struct{}

func NewInboxEventRepository() InboxEventRepositoryInterface {
	return &InboxEventRepository{}
}

func (r *InboxEventRepository) CreateOne(
	input inputs.CreateInboxEventInput,
	parsedOptions RepositoryOptionFields,
) (bool, *cexceptions.Exception) {
	if input.EventId == uuid.Nil {
		return false, cexceptions.New("InvalidInput", "InboxEvent", "Create", "Inbox event ID is required", http.StatusBadRequest)
	}
	if parsedOptions.DB == nil {
		return false, cexceptions.New(
			"DatabaseUnavailable",
			"InboxEvent",
			"Create",
			"A database connection is required",
			http.StatusInternalServerError,
			true,
		)
	}

	result := parsedOptions.DB.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&schemas.InboxEvent{EventId: input.EventId})
	if result.Error != nil {
		return false, cexceptions.New(
			"FailedToCreate",
			"InboxEvent",
			"Create",
			"Failed to record inbox event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return result.RowsAffected > 0, nil
}
