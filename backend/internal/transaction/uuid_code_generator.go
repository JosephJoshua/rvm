package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

type UUIDCodeGenerator struct{}

func NewUUIDCodeGenerator() UUIDCodeGenerator {
	return UUIDCodeGenerator{}
}

func (cg UUIDCodeGenerator) Generate() (string, error) {
	code, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("Generate(): failed to generate uuid: %w", err)
	}

	return code.String(), nil
}
