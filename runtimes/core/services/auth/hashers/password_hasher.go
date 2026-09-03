package hashers

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type PasswordHasherInterface interface {
	Hash(password string) (string, *cexceptions.Exception)
	Compare(hashedPassword string, password string) bool
}

type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasherInterface {
	return &PasswordHasher{}
}

func (h *PasswordHasher) Hash(password string) (string, *cexceptions.Exception) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", cexceptions.New(
			"FailedToGenerateHashValue",
			"Auth",
			"Hash",
			"Failed to generate a hash value",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return string(bytes), nil
}

func (h *PasswordHasher) Compare(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
