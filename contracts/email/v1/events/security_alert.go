package emaileventscontract

import "time"

type SendSecurityAlertEmailRequestDto = SendEmailRequestDto[SecurityAlertEmailPattern]

type SecurityAlertEmailPattern struct {
	UserName         string    `json:"userName" validate:"required"`
	Status           string    `json:"status" validate:"required"`
	AlertType        string    `json:"alertType" validate:"required"`
	Reason           string    `json:"reason" validate:"required"`
	TimeOfOccurrence time.Time `json:"timeOfOccurrence" validate:"required"`
	OtherDetails     string    `json:"otherDetails"`
}
