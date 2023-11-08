package transaction

import (
	"fmt"

	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
)

var (
	ErrTransactionDoesNotExist = fmt.Errorf("transaction does not exist")
	ErrItemDoesNotExist        = fmt.Errorf("item does not exist")
	ErrUserDoesNotExist        = fmt.Errorf("user does not exist")
)

type Service struct {
	r  Repository
	ig IDGenerator
}

func NewService(r Repository, cg IDGenerator) *Service {
	return &Service{r: r, ig: cg}
}

func (s *Service) StartTransaction() (domain.TransactionID, error) {
	id, err := s.ig.Generate()
	if err != nil {
		return "", fmt.Errorf("StartTransaction(): failed to generate id: %w", err)
	}

	if err = s.r.StartTransaction(id); err != nil {
		return "", fmt.Errorf("StartTransaction(): failed to create transaction: %w", err)
	}

	return id, nil
}

func (s *Service) AddItemToTransaction(transactionID domain.TransactionID, itemID int) error {
	ok, err := s.r.DoesTransactionExist(transactionID)
	if err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to check transaction existence: %w", err)
	}

	if !ok {
		return fmt.Errorf("AddItemToTransaction(): %w with id %s", ErrTransactionDoesNotExist, transactionID.String())
	}

	ok, err = s.r.DoesItemExist(itemID)
	if err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to check item existence: %w", err)
	}

	if !ok {
		return fmt.Errorf("AddItemToTransaction(): %w with id %v", ErrItemDoesNotExist, itemID)
	}

	if err = s.r.AddItemToTransaction(transactionID, itemID); err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to add item to transaction: %w", err)
	}

	return nil
}

func (s *Service) EndTransactionAndAssignUser(transactionID domain.TransactionID, userID string) error {
	ok, err := s.r.DoesTransactionExist(transactionID)
	if err != nil {
		return fmt.Errorf("EndTransactionAndAssignUser(): failed to check transaction existence: %w", err)
	}

	if !ok {
		return fmt.Errorf("EndTransactionAndAssignUser(): %w with id %s", ErrTransactionDoesNotExist, transactionID.String())
	}

	ok, err = s.r.DoesUserExist(userID)
	if err != nil {
		return fmt.Errorf("EndTransactionAndAssignUser(): failed to check user existence: %w", err)
	}

	if !ok {
		return fmt.Errorf("EndTransactionAndAssignUser(): %w with id %s", ErrUserDoesNotExist, userID)
	}

	if err = s.r.EndTransactionAndAssignUser(transactionID, userID); err != nil {
		return fmt.Errorf("EndTransactionAndAssignUser(): failed to end transaction: %w", err)
	}

	return nil
}
