package auth

import (
	"fmt"

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

func (ar *SQLRepository) DoesUserExist(uid string) (bool, error) {
	var count int
	if err := ar.db.Get(&count, `
		SELECT
			COUNT(*)
		FROM
			users
		WHERE
			user_id = ?
	`, uid); err != nil {
		return false, fmt.Errorf("GetTransactionItemCount(): failed to execute query: %w", err)
	}

	return count > 0, nil
}

func (ar *SQLRepository) CreateUser(uid string, fullName string, email string) error {
	if _, err := ar.db.Exec(`
		INSERT INTO
			users (user_id, full_name, email)
		VALUES
			(?, ?, ?)
	`, uid, fullName, email); err != nil {
		return fmt.Errorf("CreateUser(): failed to execute query: %w", err)
	}

	return nil
}
