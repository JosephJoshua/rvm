package domain

import "time"

type Transaction struct {
	ID        TransactionID
	UserID    string
	CreatedAt time.Time
}

func NewTransaction(id TransactionID, userID string, createdAt time.Time) Transaction {
	return Transaction{
		ID:        id,
		UserID:    userID,
		CreatedAt: createdAt,
	}
}
