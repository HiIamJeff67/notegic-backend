package routinetasktypes

import (
	"github.com/google/uuid"
)

type PrecreatedSubShelf struct {
	TaskId           uuid.UUID                   `json:"taskId" validate:"required"`
	FakeId           RoutineTaskObjectReference  `json:"fakeId" validate:"required"`
	RealId           uuid.UUID                   `json:"realId" validate:"required"`
	RootShelfId      uuid.UUID                   `json:"rootShelfId" validate:"required"`
	ParentSubShelfId *RoutineTaskObjectReference `json:"parentSubShelfId,omitempty"`
	Name             string                      `json:"name" validate:"required"`
	Path             []uuid.UUID                 `json:"path"`
}

type RoutineTaskPlan struct {
	RoutineId               uuid.UUID                     `json:"routineId" validate:"required"`
	Facts                   map[string]uuid.UUID          `json:"facts" validate:"required"`
	PrecreatedSubShelves    map[string]PrecreatedSubShelf `json:"precreatedSubShelves" validate:"dive"`
	PrecreatedSubShelfOrder []string                      `json:"precreatedSubShelfOrder"`
	ContainerObjectTaskIds  []uuid.UUID                   `json:"containerObjectTaskIds" validate:"dive"`
	CoreObjectTaskIds       []uuid.UUID                   `json:"coreObjectTaskIds" validate:"dive"`
	PlannedObjectIds        map[string]uuid.UUID          `json:"plannedObjectIds" validate:"dive"`
}
