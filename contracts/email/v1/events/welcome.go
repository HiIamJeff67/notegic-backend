package emaileventscontract

type SendWelcomeEmailRequestDto = SendEmailRequestDto[WelcomeEmailPattern]

type WelcomeEmailPattern struct {
	UserName     string                           `json:"userName" validate:"required"`
	Email        string                           `json:"email" validate:"omitempty,email"`
	Status       string                           `json:"status" validate:"required"`
	RoutineItems []WelcomeEmailRoutineItemPattern `json:"routineItems" validate:"omitempty,dive"`
}

type WelcomeEmailRoutineItemPattern struct {
	Name   string `json:"name" validate:"required"`
	Status string `json:"status" validate:"required"`
}

type WelcomeEmailRoutineItem = WelcomeEmailRoutineItemPattern
