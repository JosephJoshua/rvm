package transaction

import (
	"database/sql"
	"fmt"

	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (tr *SQLRepository) DoesTransactionExist(id domain.TransactionID) (bool, error) {
	stmt, err := tr.db.Prepare(`
		SELECT
			COUNT(*)
		FROM
			transactions
		WHERE
			transaction_id = ?
	`)

	if err != nil {
		return false, fmt.Errorf("DoesTransactionExist(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	var count int
	if err = stmt.QueryRow(id).Scan(&count); err != nil {
		return false, fmt.Errorf("DoesTransactionExist(): failed to execute statement: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) DoesItemExist(itemID int) (bool, error) {
	stmt, err := tr.db.Prepare(`
		SELECT
			COUNT(*)
		FROM
			items
		WHERE
			item_id = ?
	`)

	if err != nil {
		return false, fmt.Errorf("DoesItemExist(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	var count int
	if err = stmt.QueryRow(itemID).Scan(&count); err != nil {
		return false, fmt.Errorf("DoesItemExist(): failed to execute statement: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) DoesUserExist(userID string) (bool, error) {
	stmt, err := tr.db.Prepare(`
		SELECT
			COUNT(*)
		FROM
			users
		WHERE
			user_id = ?
	`)

	if err != nil {
		return false, fmt.Errorf("DoesUserExist(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	var count int
	if err = stmt.QueryRow(userID).Scan(&count); err != nil {
		return false, fmt.Errorf("DoesUserExist(): failed to execute statement: %w", err)
	}

	return count > 0, nil
}

func (tr *SQLRepository) StartTransaction(id domain.TransactionID) error {
	stmt, err := tr.db.Prepare(`
		INSERT INTO
			transactions (transaction_id)
		VALUES
			(?)
	`)

	if err != nil {
		return fmt.Errorf("StartTransaction(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	if _, err = stmt.Exec(id); err != nil {
		return fmt.Errorf("StartTransaction(): failed to execute statement: %w", err)
	}

	return nil
}

func (tr *SQLRepository) AddItemToTransaction(transactionID domain.TransactionID, itemID int) error {
	stmt, err := tr.db.Prepare(`
		INSERT INTO
			transaction_items (transaction_id, item_id)
		VALUES
			(?, ?)
	`)

	if err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	if _, err = stmt.Exec(transactionID, itemID); err != nil {
		return fmt.Errorf("AddItemToTransaction(): failed to execute statement: %w", err)
	}

	return nil
}

func (tr *SQLRepository) EndTransactionAndAssignUser(transactionID domain.TransactionID, userID string) error {
	stmt, err := tr.db.Prepare(`
		UPDATE
			transactions
		SET
			user_id = ?
		WHERE
			transaction_id = ?
	`)

	if err != nil {
		return fmt.Errorf("EndTransactionAndAssignUser(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	if _, err = stmt.Exec(userID, transactionID); err != nil {
		return fmt.Errorf("EndTransactionAndAssignUser(): failed to execute statement: %w", err)
	}

	return nil
}
