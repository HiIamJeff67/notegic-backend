package routinetasktypes

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const RoutineTaskFakeIdPrefix = "f_"

type RoutineTaskObjectReference string

func NewRoutineTaskFakeId() RoutineTaskObjectReference {
	return RoutineTaskObjectReference(RoutineTaskFakeIdPrefix + strings.ReplaceAll(uuid.NewString(), "-", ""))
}

func (reference RoutineTaskObjectReference) IsFakeId() bool {
	value := string(reference)
	if !strings.HasPrefix(value, RoutineTaskFakeIdPrefix) {
		return false
	}

	encodedId := strings.TrimPrefix(value, RoutineTaskFakeIdPrefix)
	if len(encodedId) != 32 {
		return false
	}

	_, err := hex.DecodeString(encodedId)
	return err == nil
}

func (reference RoutineTaskObjectReference) IsRealId() bool {
	_, err := uuid.Parse(string(reference))
	return err == nil
}

func (reference RoutineTaskObjectReference) Validate() error {
	if reference.IsFakeId() || reference.IsRealId() {
		return nil
	}

	return fmt.Errorf("routine task object reference %q is neither a fake id nor a uuid", reference)
}

func (reference RoutineTaskObjectReference) Resolve(facts map[string]uuid.UUID) (uuid.UUID, error) {
	if reference.IsFakeId() {
		resolvedId, ok := facts[string(reference)]
		if !ok {
			return uuid.Nil, fmt.Errorf("routine task fake id %q is not present in facts", reference)
		}
		return resolvedId, nil
	}

	resolvedId, err := uuid.Parse(string(reference))
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse routine task object reference %q: %w", reference, err)
	}
	return resolvedId, nil
}
