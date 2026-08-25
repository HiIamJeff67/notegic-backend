package inputs

type PartialUpdateInput[T any] struct {
	Values  T                `json:"values"`
	SetNull *map[string]bool `json:"setNull"`
}
