package emaileventscontract

import "time"

type SendValidationEmailRequestDto = SendEmailRequestDto[ValidationEmailPattern]

type ValidationEmailPattern struct {
	UserName      string    `json:"userName" validate:"required"`
	Email         string    `json:"email" validate:"omitempty,email"`
	AuthCode      string    `json:"authCode" validate:"required"`
	UserAgent     string    `json:"userAgent" validate:"required"`
	ExpiredAt     time.Time `json:"expiredAt" validate:"required"`
	ExpiryMinutes int       `json:"expiryMinutes"`
	RequestTime   string    `json:"requestTime"`
}
