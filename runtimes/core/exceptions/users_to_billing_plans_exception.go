package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UsersToBillingPlansException struct {
	Exception
}

func NewUsersToBillingPlansException() UsersToBillingPlansException {
	return UsersToBillingPlansException{Exception: NewException("UsersToBillingPlans")}
}

func (UsersToBillingPlansException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"UsersToBillingPlans",
		"Repository",
		"UsersToBillingPlans was not found",
		http.StatusNotFound,
	)
}

func (UsersToBillingPlansException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"UsersToBillingPlans",
		"Repository",
		"Failed to create UsersToBillingPlans",
		http.StatusInternalServerError,
		true,
	)
}

func (UsersToBillingPlansException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"UsersToBillingPlans",
		"Repository",
		"Failed to update UsersToBillingPlans",
		http.StatusInternalServerError,
		true,
	)
}

func (UsersToBillingPlansException) FailedToDelete() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToDelete",
		"UsersToBillingPlans",
		"Repository",
		"Failed to delete UsersToBillingPlans",
		http.StatusInternalServerError,
		true,
	)
}

func (UsersToBillingPlansException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"UsersToBillingPlans",
		"Repository",
		"No changes were applied to UsersToBillingPlans",
		http.StatusNotModified,
	)
}
