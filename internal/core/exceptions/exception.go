package apiexceptions

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

// Exception contains exception builders owned by this runtime's services, workers, and handlers.
type Exception struct {
	Domain string
}

func NewException(domain string) Exception {
	return Exception{Domain: domain}
}

func (e Exception) New(
	reason string,
	operation string,
	message string,
	status int,
	retryable ...bool,
) *cexceptions.Exception {
	return cexceptions.New(reason, e.Domain, operation, message, status, retryable...)
}

func (e Exception) NewForDomain(
	reason string,
	domain string,
	operation string,
	message string,
	status int,
	retryable ...bool,
) *cexceptions.Exception {
	return cexceptions.New(reason, domain, operation, message, status, retryable...)
}

func (e Exception) NotFound(optionalMessage ...string) *cexceptions.Exception {
	message := e.Domain + " was not found"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("NotFound", "Repository", message, http.StatusNotFound)
}

func (e Exception) FailedToGet(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to get " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToGet", "Repository", message, http.StatusInternalServerError, true)
}

func (e Exception) FailedToCreate(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to create " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToCreate", "Repository", message, http.StatusInternalServerError, true)
}

func (e Exception) FailedToUpdate(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to update " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToUpdate", "Repository", message, http.StatusInternalServerError, true)
}

func (e Exception) FailedToDelete(optionalMessage ...string) *cexceptions.Exception {
	message := "Failed to delete " + e.Domain
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("FailedToDelete", "Repository", message, http.StatusInternalServerError, true)
}

func (e Exception) NoChanges() *cexceptions.Exception {
	return e.New(
		"NoChanges",
		"Repository",
		"No changes were applied to "+e.Domain,
		http.StatusNotModified,
	)
}

func (e Exception) FailedToCommitTransaction() *cexceptions.Exception {
	return e.New(
		"FailedToCommitTransaction",
		"Transaction",
		"Failed to commit the "+e.Domain+" transaction",
		http.StatusInternalServerError,
		true,
	)
}

func (e Exception) NoPermission(action string) *cexceptions.Exception {
	return e.New(
		"PermissionDenied",
		"Authorize",
		"Permission is denied to "+action,
		http.StatusForbidden,
	)
}

func (e Exception) InvalidInput(optionalMessage ...string) *cexceptions.Exception {
	message := "Invalid " + e.Domain + " input"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("InvalidInput", "Validate", message, http.StatusBadRequest)
}

func (e Exception) InvalidDto(optionalMessage ...string) *cexceptions.Exception {
	message := "Invalid " + e.Domain + " DTO"
	if len(optionalMessage) > 0 && strings.TrimSpace(optionalMessage[0]) != "" {
		message = optionalMessage[0]
	}

	return e.New("InvalidDto", "Validate", message, http.StatusBadRequest)
}

func (e Exception) InvalidType(value any) *cexceptions.Exception {
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

func (e Exception) FailedToCompileRegularExpression() *cexceptions.Exception {
	return e.New(
		"FailedToCompileRegularExpression",
		"Validate",
		"Failed to compile regular expression",
		http.StatusInternalServerError,
		true,
	)
}

func (e Exception) CannotGetFileObjects() *cexceptions.Exception {
	return e.New(
		"CannotGetFileObjects",
		"File",
		"Failed to get file objects",
		http.StatusInternalServerError,
		true,
	)
}

func (e Exception) FailedToMarshalData(data any) *cexceptions.Exception {
	return e.New(
		"FailedToMarshal",
		"Marshal",
		fmt.Sprintf("Failed to marshal data of %v", data),
		http.StatusInternalServerError,
		true,
	)
}

func (e Exception) DatabaseUnavailable() *cexceptions.Exception {
	return e.New(
		"DatabaseUnavailable",
		"Repository",
		"A database connection is required",
		http.StatusInternalServerError,
		true,
	)
}

func (e Exception) TransactionRequired() *cexceptions.Exception {
	return e.New(
		"TransactionRequired",
		"Create",
		e.Domain+" operations must be created in the domain transaction",
		http.StatusInternalServerError,
	)
}

func (e Exception) DuplicateName(name string) *cexceptions.Exception {
	return e.New(
		"DuplicateName",
		"Create",
		fmt.Sprintf("The name of %s is already in use", name),
		http.StatusConflict,
	)
}

func (e Exception) DuplicateEmail(email string) *cexceptions.Exception {
	return e.New(
		"DuplicateEmail",
		"Create",
		fmt.Sprintf("The email %s is already in use", email),
		http.StatusConflict,
	)
}

func (e Exception) NoRootBlockInBlockPack(blockPackId uuid.UUID) *cexceptions.Exception {
	return e.New(
		"NoRootBlockInBlockPack",
		"Project",
		fmt.Sprintf("No root block exists in block pack %s", blockPackId),
		http.StatusInternalServerError,
		true,
	)
}
