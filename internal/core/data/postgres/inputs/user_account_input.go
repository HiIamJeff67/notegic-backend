package inputs

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type CreateUserAccountInput struct {
	AuthCode          string             `json:"authCode" gorm:"column:auth_code"`
	AuthCodeExpiredAt time.Time          `json:"authCodeExpiredAt" gorm:"column:auth_code_expired_at"`
	CountryCode       *enums.CountryCode `json:"countryCode" gorm:"column:country_code;"`
	BackupEmail       *string            `json:"backupEmail" gorm:"column:backup_email;"`
	PhoneNumber       *string            `json:"phoneNumber" gorm:"column:phone_number;"`
	GoogleCredential  *string            `json:"googleCredential" gorm:"column:google_credential;"`
	DiscordCredential *string            `json:"discordCredential" gorm:"column:discord_credential;"`
}

type UpdateUserAccountInput struct {
	AuthCode           *string            `json:"authCode" gorm:"column:auth_code"`
	AuthCodeExpiredAt  *time.Time         `json:"authCodeExpiredAt" gorm:"column:auth_code_expired_at"`
	BlockAuthCodeUntil *time.Time         `json:"blockAuthCodeUntil" gorm:"column:block_auth_code_until"`
	CountryCode        *enums.CountryCode `json:"countryCode" gorm:"column:country_code;"`
	BackupEmail        *string            `json:"backupEmail" gorm:"column:backup_email;"`
	PhoneNumber        *string            `json:"phoneNumber" gorm:"column:phone_number;"`
	GoogleCredential   *string            `json:"googleCredential" gorm:"column:google_credential;"`
	DiscordCredential  *string            `json:"discordCredential" gorm:"column:discord_credential;"`
}

type PartialUpdateUserAccountInput = PartialUpdateInput[UpdateUserAccountInput]
