package routinetasktypes

import "github.com/google/uuid"

type GetSubShelfRoutineTaskPayload struct {
	SubShelfId uuid.UUID `json:"subShelfId" validate:"required"`
}

type DeleteSubShelfRoutineTaskPayload struct {
	SubShelfId uuid.UUID `json:"subShelfId" validate:"required"`
}

type CreateSubShelfRoutineTaskPayload struct {
	Id             *uuid.UUID                  `json:"id" validate:"omitnil"`
	FakeId         RoutineTaskObjectReference  `json:"fakeId" validate:"required"`
	RootShelfId    uuid.UUID                   `json:"rootShelfId" validate:"required"`
	PrevSubShelfId *RoutineTaskObjectReference `json:"prevSubShelfId" validate:"omitnil"`
	Path           []uuid.UUID                 `json:"path,omitempty" validate:"omitempty,max=100"`
	Name           string                      `json:"name" validate:"required,min=1,max=128,isshelfname"`
	Pattern        RoutineTaskPattern          `json:"pattern" validate:"omitempty,dive"`
}

type UpdateSubShelfRoutineTaskPayload struct {
	SubShelfId uuid.UUID          `json:"subShelfId" validate:"required"`
	Name       *string            `json:"name" validate:"omitnil,min=1,max=128,isshelfname"`
	Pattern    RoutineTaskPattern `json:"pattern" validate:"omitempty,dive"`
}
