package transaction

import (
	"fmt"

	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
	"github.com/google/uuid"
)

type UUIDIDGenerator struct{}

func NewUUIDIDGenerator() UUIDIDGenerator {
	return UUIDIDGenerator{}
}

func (cg UUIDIDGenerator) Generate() (domain.TransactionID, error) {
	idStr, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("Generate(): failed to generate uuid: %w", err)
	}

	id, err := domain.NewTransactionID(idStr.String())
	if err != nil {
		return "", fmt.Errorf("Generate(): failed to create transaction id: %w", err)
	}

	return id, nil
}
