package domain

import (
	"fmt"
)

const (
	transactionIDMinLength = 12
	transactionIDMaxLength = 256
)

type TransactionID string

func NewTransactionID(value string) (TransactionID, error) {
	if len(value) < transactionIDMinLength {
		return "", fmt.Errorf("transaction id must be at least %d characters long", transactionIDMinLength)
	}

	if len(value) > transactionIDMaxLength {
		return "", fmt.Errorf("transaction id must be at most %d characters long", transactionIDMaxLength)
	}

	// We don't bother checking the uniqueness of this transaction id
	// to not add any unnecessary complexity.
	// NOTE: This is not good practice in production.
	//       Maybe implement a check for uniqueness in the future
	//       by passing in a `TransactionIDUniquenessChecker` interface.

	return TransactionID(value), nil
}

func (id TransactionID) String() string {
	return string(id)
}
