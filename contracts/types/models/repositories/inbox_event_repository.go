package repositories

import (
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	models "github.com/HiIamJeff67/notegic-backend/contracts/types/models"
	inputs "github.com/HiIamJeff67/notegic-backend/contracts/types/models/inputs"
)

type InboxEventRepositoryInterface interface {
	CreateOne(input inputs.CreateInboxEventInput, opts RepositoryOptionFields) (bool, *exceptions.Exception)
}

type InboxEventRepository struct{}

func NewInboxEventRepository() InboxEventRepositoryInterface {
	return &InboxEventRepository{}
}

func (r *InboxEventRepository) CreateOne(
	input inputs.CreateInboxEventInput,
	parsedOptions RepositoryOptionFields,
) (bool, *exceptions.Exception) {
	if input.EventId == uuid.Nil {
		return false, exceptions.New("InvalidInput", "InboxEvent", "Create", "Inbox event ID is required", http.StatusBadRequest)
	}
	if parsedOptions.DB == nil {
		return false, exceptions.New(
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
		Create(&models.InboxEvent{EventId: input.EventId})
	if result.Error != nil {
		return false, exceptions.New(
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
