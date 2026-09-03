package generators

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
)

type FakeDisplayNameGeneratorInterface interface {
	GenerateRandomly() string
}

type FakeDisplayNameGenerator struct{}

func NewFakeDisplayNameGenerator() FakeDisplayNameGeneratorInterface {
	return &FakeDisplayNameGenerator{}
}

func (g *FakeDisplayNameGenerator) GenerateRandomly() string {
	gofakeit.Seed(0)
	return fmt.Sprintf(
		"%s%s%d",
		gofakeit.AdjectiveDescriptive(),
		gofakeit.LastName(),
		gofakeit.Number(100000, 999999),
	)
}
