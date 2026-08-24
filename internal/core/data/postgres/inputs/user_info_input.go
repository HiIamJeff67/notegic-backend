package inputs

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type CreateUserInfoInput struct {
	CoverBackgroundURL *string           `json:"coverBackgroundURL" gorm:"column:cover_background_url;"`
	AvatarURL          *string           `json:"avatarURL" gorm:"column:avatar_url;"`
	Header             *string           `json:"header" gorm:"column:header;"`
	Introduction       *string           `json:"introduction" gorm:"column:introduction;"`
	Gender             *enums.UserGender `json:"gender" gorm:"column:gender;"`
	Country            *enums.Country    `json:"country" gorm:"column:country;"`
	BirthDate          *time.Time        `json:"birthDate" gorm:"column:birth_date;"`
}

type UpdateUserInfoInput struct {
	CoverBackgroundURL *string           `json:"coverBackgroundURL" gorm:"column:cover_background_url;"`
	AvatarURL          *string           `json:"avatarURL" gorm:"column:avatar_url;"`
	Header             *string           `json:"header" gorm:"column:header;"`
	Introduction       *string           `json:"introduction" gorm:"column:introduction;"`
	Gender             *enums.UserGender `json:"gender" gorm:"column:gender;"`
	Country            *enums.Country    `json:"country" gorm:"column:country;"`
	BirthDate          *time.Time        `json:"birthDate" gorm:"column:birth_date;"`
}

type PartialUpdateUserInfoInput = PartialUpdateInput[UpdateUserInfoInput]
