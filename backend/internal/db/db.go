package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	// sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("NewDB(): failed to open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("NewDB(): failed to ping db: %w", err)
	}

	return db, nil
}

func MigrateDB(db *sqlx.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS api_tokens (
			api_token_id VARCHAR(255) PRIMARY KEY NOT NULL,
			expiring_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL
		) WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate api_tokens: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL,
			birth_date DATE NOT NULL
		) WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate users: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			transaction_id VARCHAR(255) PRIMARY KEY NOT NULL,
			user_id INTEGER NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE ON UPDATE CASCADE
		) WITHOUT ROWID;
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate transactions: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			item_id INTEGER PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			points int NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate items: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transaction_items (
			transaction_item_id INTEGER PRIMARY KEY NOT NULL,
			transaction_id INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (transaction_id) REFERENCES transactions (transaction_id) ON DELETE CASCADE ON UPDATE CASCADE,
			FOREIGN KEY (item_id) REFERENCES items (item_id) ON DELETE CASCADE ON UPDATE CASCADE
		);
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate transaction_items: %w", err)
	}

	return nil
}

func SeedDB(db *sqlx.DB) error {
	if _, err := db.Exec(`
		INSERT INTO
			items (name, points)
		VALUES
			('PET bottle', 10)
	`); err != nil {
		return fmt.Errorf("SeedDB(): failed to seed items: %w", err)
	}

	return nil
}
