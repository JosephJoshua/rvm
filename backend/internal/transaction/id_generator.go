package transaction

import "github.com/JosephJoshua/rvm/backend/internal/transaction/domain"

type IDGenerator interface {
	Generate() (domain.TransactionID, error)
}
