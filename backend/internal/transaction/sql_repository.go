package transaction

import (
	"database/sql"
	"fmt"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (tr *SQLRepository) StartTransaction(code string) error {
	stmt, err := tr.db.Prepare(`INSERT INTO transactions (transaction_code) VALUES (?)`)
	if err != nil {
		return fmt.Errorf("StartTransaction(): failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	if _, err = stmt.Exec(code); err != nil {
		return fmt.Errorf("StartTransaction(): failed to execute statement: %w", err)
	}

	return nil
}
