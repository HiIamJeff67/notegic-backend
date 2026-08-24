package inputs

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type CreateUserInput struct {
	Name         string `json:"name" gorm:"column:name;"`
	DisplayName  string `json:"displayName" gorm:"column:display_name"`
	Email        string `json:"email" gorm:"column:email;"`
	Password     string `json:"password" gorm:"column:password;"` // hashed password
	RefreshToken string `json:"refreshToken" gorm:"column:refresh_token;"`
	UserAgent    string `json:"userAgent" gorm:"column:user_agent;"`
}

type UpdateUserInput struct {
	Name           *string           `json:"name" gorm:"column:name;"`
	DisplayName    *string           `json:"displayName" gorm:"column:display_name;"`
	Email          *string           `json:"email" gorm:"column:email;"`
	Password       *string           `json:"password" gorm:"column:password;"`
	RefreshToken   *string           `json:"refreshToken" gorm:"column:refresh_token;"`
	LoginCount     *int32            `json:"loginCount" gorm:"column:login_count;"`
	BlockLoginUtil *time.Time        `json:"blockLoginUntil" gorm:"column:block_login_until"`
	UserAgent      *string           `json:"userAgent" gorm:"column:user_agent;"`
	Role           *enums.UserRole   `json:"role" gorm:"column:role;"`
	Plan           *enums.UserPlan   `json:"plan" gorm:"column:plan;"`
	Status         *enums.UserStatus `json:"status" gorm:"column:status;"`
}

type PartialUpdateUserInput = PartialUpdateInput[UpdateUserInput]
