package exceptions

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

// DurableJobException contains exception builders owned by this runtime's services, workers, and handlers.
type DurableJobException struct {
	Domain string
}

func NewDurableJobException() DurableJobException {
	return DurableJobException{
		Domain: "DurableJob",
	}
}

func (e DurableJobException) New(
	reason string,
	operation string,
	message string,
	status int,
	retryable ...bool,
) *cexceptions.Exception {
	return cexceptions.New(reason, e.Domain, operation, message, status, retryable...)
}

func (e DurableJobException) NotFound(optionalMessage ...string) *cexceptions.Exception {
	message := e.Domain + " was not found"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("NotFound", "Repository", message, http.StatusNotFound)
}

func (e DurableJobException) FailedToGet(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to get " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToGet", "Repository", message, http.StatusInternalServerError, true)
}

func (e DurableJobException) FailedToCreate(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to create " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToCreate", "Repository", message, http.StatusInternalServerError, true)
}

func (e DurableJobException) FailedToUpdate(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to update " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToUpdate", "Repository", message, http.StatusInternalServerError, true)
}

func (e DurableJobException) FailedToDelete(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to delete " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToDelete", "Repository", message, http.StatusInternalServerError, true)
}

func (e DurableJobException) NoChanges() *cexceptions.Exception {
	return e.New(
		"NoChanges",
		"Repository",
		"No changes were applied to "+e.Domain,
		http.StatusNotModified,
	)
}

func (e DurableJobException) FailedToCommitTransaction() *cexceptions.Exception {
	return e.New(
		"FailedToCommitTransaction",
		"Transaction",
		"Failed to commit the "+e.Domain+" transaction",
		http.StatusInternalServerError,
		true,
	)
}

func (e DurableJobException) NoPermission(action string) *cexceptions.Exception {
	return e.New(
		"PermissionDenied",
		"Authorize",
		"Permission is denied to "+action,
		http.StatusForbidden,
	)
}

func (e DurableJobException) InvalidInput(optionalMessage ...string) *cexceptions.Exception {
	message := "Invalid " + e.Domain + " input"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("InvalidInput", "Validate", message, http.StatusBadRequest)
}

func (e DurableJobException) InvalidDto(optionalMessage ...string) *cexceptions.Exception {
	message := "Invalid " + e.Domain + " DTO"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("InvalidDto", "Validate", message, http.StatusBadRequest)
}

func (e DurableJobException) InvalidType(value any) *cexceptions.Exception {
	return e.New(
		"InvalidType",
		"Validate",
		"Invalid type in "+e.Domain,
		http.StatusInternalServerError,
		true,
	).WithDetails(map[string]any{
		"actualType": fmt.Sprintf("%T", value),
		"value":      value,
	})
}

func (e DurableJobException) FailedToCompileRegularExpression() *cexceptions.Exception {
	return e.New(
		"FailedToCompileRegularExpression",
		"Validate",
		"Failed to compile regular expression",
		http.StatusInternalServerError,
		true,
	)
}

func (e DurableJobException) CannotGetFileObjects() *cexceptions.Exception {
	return e.New(
		"CannotGetFileObjects",
		"File",
		"Failed to get file objects",
		http.StatusInternalServerError,
		true,
	)
}

func (e DurableJobException) FailedToMarshalData(data any) *cexceptions.Exception {
	return e.New(
		"FailedToMarshal",
		"Marshal",
		fmt.Sprintf("Failed to marshal data of %v", data),
		http.StatusInternalServerError,
		true,
	)
}

func (e DurableJobException) DatabaseUnavailable() *cexceptions.Exception {
	return e.New(
		"DatabaseUnavailable",
		"Repository",
		"A database connection is required",
		http.StatusInternalServerError,
		true,
	)
}

func (e DurableJobException) TransactionRequired() *cexceptions.Exception {
	return e.New(
		"TransactionRequired",
		"Create",
		e.Domain+" operations must be created in the domain transaction",
		http.StatusInternalServerError,
	)
}

func (e DurableJobException) DuplicateName(name string) *cexceptions.Exception {
	return e.New(
		"DuplicateName",
		"Create",
		fmt.Sprintf("The name of %s is already in use", name),
		http.StatusConflict,
	)
}

func (e DurableJobException) DuplicateEmail(email string) *cexceptions.Exception {
	return e.New(
		"DuplicateEmail",
		"Create",
		fmt.Sprintf("The email %s is already in use", email),
		http.StatusConflict,
	)
}

func (e DurableJobException) NoRootBlockInBlockPack(blockPackId uuid.UUID) *cexceptions.Exception {
	return e.New(
		"NoRootBlockInBlockPack",
		"Project",
		fmt.Sprintf("No root block exists in block pack %s", blockPackId),
		http.StatusInternalServerError,
		true,
	)
}
