package transaction

import (
	"fmt"
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
	"github.com/jmoiron/sqlx"
)

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (tr *SQLRepository) DoesTransactionExist(transactionID domain.TransactionID) (bool, error) {
	var count int
	if err := tr.db.Get(&count, `
		SELECT
			COUNT(*)
		FROM
			transactions
		WHERE
			transaction_id = ?
	`, transactionID); err != nil {
		return false, fmt.Errorf("DoesTransactionExist(): failed to execute query: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) DoesItemExist(itemID int) (bool, error) {
	var count int
	if err := tr.db.Get(&count, `
		SELECT
			COUNT(*)
		FROM
			items
		WHERE
			item_id = ?
	`, itemID); err != nil {
		return false, fmt.Errorf("DoesItemExist(): failed to execute query: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) DoesUserExist(userID string) (bool, error) {
	var count int
	if err := tr.db.Get(&count, `
		SELECT
			COUNT(*)
		FROM
			users
		WHERE
			user_id = ?
	`, userID); err != nil {
		return false, fmt.Errorf("DoesUserExist(): failed to execute query: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) StartTransaction(id domain.TransactionID, createdAt time.Time) error {
	if _, err := tr.db.Exec(`
		INSERT INTO
			transactions (transaction_id, created_at)
		VALUES
			(?, ?)
	`, id, createdAt); err != nil {
		return fmt.Errorf("StartTransaction(): failed to execute query: %w", err)
	}

	return nil
}

func (tr *SQLRepository) AddItemToTransaction(
	transactionID domain.TransactionID,
	itemID int,
	createdAt time.Time,
) error {
	if _, err := tr.db.Exec(`
		INSERT INTO
			transaction_items (transaction_id, item_id, created_at)
		VALUES
			(?, ?, ?)
	`, transactionID, itemID, createdAt); err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to execute query: %w", err)
	}

	return nil
}

func (tr *SQLRepository) EndTransactionAndAssignUser(transactionID domain.TransactionID, userID string) error {
	if _, err := tr.db.Exec(`
		UPDATE
			transactions
		SET
			user_id = ?
		WHERE
			transaction_id = ?
	`, userID, transactionID); err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to execute query: %w", err)
	}

	return nil
}
