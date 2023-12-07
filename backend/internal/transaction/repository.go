package transaction

import (
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
)

type Repository interface {
	DoesTransactionExist(id domain.TransactionID) (bool, error)
	DoesItemExist(itemID int) (bool, error)
	DoesUserExist(userID string) (bool, error)
	StartTransaction(id domain.TransactionID, createdAt time.Time) error
	AddItemToTransaction(transactionID domain.TransactionID, itemID int, createdAt time.Time) error
	EndTransactionAndAssignUser(transactionID domain.TransactionID, userID string) error
	IsTransactionAssigned(transactionID domain.TransactionID) (bool, error)
	GetTransactionItemCount(transactionID domain.TransactionID) (int, error)
	GetTransactionPoints(transactionID domain.TransactionID) (int, error)
}
