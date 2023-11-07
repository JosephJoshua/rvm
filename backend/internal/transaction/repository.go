package transaction

import (
	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
)

type Repository interface {
	DoesTransactionExist(id domain.TransactionID) (bool, error)
	DoesItemExist(itemID int) (bool, error)
	StartTransaction(id domain.TransactionID) error
	AddItemToTransaction(transactionID domain.TransactionID, itemID int) error
}
