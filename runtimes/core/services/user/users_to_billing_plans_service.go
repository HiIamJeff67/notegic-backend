package user

import (
	"gorm.io/gorm"
)

type UsersToBillingPlansServiceInterface interface{}

type UsersToBillingPlansService struct {
	db *gorm.DB
	// usersToBillingPlansRepository repositories.usersToBillingPlansRepository
}
