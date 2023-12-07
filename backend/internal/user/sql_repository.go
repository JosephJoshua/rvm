package user

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (ur *SQLRepository) GetPoints(uid string) (int, error) {
	var points int
	if err := ur.db.Get(&points, `
		SELECT
			COALESCE(SUM(items.points), 0)
		FROM
			users
		LEFT JOIN
			transactions ON transactions.user_id = users.user_id
		LEFT JOIN
			transaction_items ON transaction_items.transaction_id = transactions.transaction_id
		LEFT JOIN
			items ON items.item_id = transaction_items.item_id
		WHERE
			users.user_id = ?
	`, uid); err != nil {
		return 0, fmt.Errorf("GetPoints(): failed to execute query: %w", err)
	}

	return points, nil
}
