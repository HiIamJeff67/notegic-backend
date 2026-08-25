package apiexceptions

import (
	"fmt"
	"net/http"
	"time"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type AuthException struct {
	Exception
}

func NewAuthException() AuthException {
	return AuthException{
		Exception: NewException("Auth"),
	}
}

func (AuthException) WrongPassword() *cexceptions.Exception {
	return cexceptions.New(
		"WrongPassword",
		"Auth",
		"Authenticate",
		"The password does not match",
		http.StatusUnauthorized,
	)
}

func (AuthException) WrongAuthCode() *cexceptions.Exception {
	return cexceptions.New(
		"WrongAuthCode",
		"Auth",
		"Authenticate",
		"The authentication code does not match",
		http.StatusUnauthorized,
	)
}

func (AuthException) LoginBlockedDueToTryingTooManyTimes(blockedUntil time.Time) *cexceptions.Exception {
	return cexceptions.New(
		"LoginBlockedDueToTryingTooManyTimes",
		"Auth",
		"Authenticate",
		fmt.Sprintf("Login is blocked until %v", blockedUntil),
		http.StatusTooManyRequests,
	)
}

func (AuthException) AuthCodeBlockedDueToTryingTooManyTimes(blockedUntil time.Time) *cexceptions.Exception {
	return cexceptions.New(
		"AuthCodeBlockedDueToTryingTooManyTimes",
		"Auth",
		"Authenticate",
		fmt.Sprintf("Auth code generation is blocked until %v", blockedUntil),
		http.StatusTooManyRequests,
	)
}
