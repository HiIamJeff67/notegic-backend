package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UserSettingException struct {
	Exception
}

func NewUserSettingException() UserSettingException {
	return UserSettingException{Exception: NewException("UserSetting")}
}

func (UserSettingException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"UserSetting",
		"Repository",
		"UserSetting was not found",
		http.StatusNotFound,
	)
}

func (UserSettingException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"UserSetting",
		"Repository",
		"Failed to create UserSetting",
		http.StatusInternalServerError,
		true,
	)
}

func (UserSettingException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"UserSetting",
		"Repository",
		"Failed to update UserSetting",
		http.StatusInternalServerError,
		true,
	)
}

func (UserSettingException) FailedToDelete() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToDelete",
		"UserSetting",
		"Repository",
		"Failed to delete UserSetting",
		http.StatusInternalServerError,
		true,
	)
}

func (UserSettingException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"UserSetting",
		"Repository",
		"No changes were applied to UserSetting",
		http.StatusNotModified,
	)
}
