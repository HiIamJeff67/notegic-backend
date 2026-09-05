package emaileventscontract

import cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"

type SendWelcomeEmailResponseDto struct {
	To               string                  `json:"to"`
	Subject          string                  `json:"subject"`
	Body             string                  `json:"body"`
	EmailContentType cemail.EmailContentType `json:"emailContentType"`
	MaxRetries       int                     `json:"maxRetries"`
	Priority         int                     `json:"priority"`
}

type SendValidationEmailResponseDto struct {
	To               string                  `json:"to"`
	Subject          string                  `json:"subject"`
	Body             string                  `json:"body"`
	EmailContentType cemail.EmailContentType `json:"emailContentType"`
	MaxRetries       int                     `json:"maxRetries"`
	Priority         int                     `json:"priority"`
}

type SendSecurityAlertEmailResponseDto struct {
	To               string                  `json:"to"`
	Subject          string                  `json:"subject"`
	Body             string                  `json:"body"`
	EmailContentType cemail.EmailContentType `json:"emailContentType"`
	MaxRetries       int                     `json:"maxRetries"`
	Priority         int                     `json:"priority"`
}
