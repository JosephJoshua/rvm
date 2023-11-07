package db

import (
	"database/sql"
	"fmt"

	// sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("NewDB(): failed to open db: %w", err)
	}

	return db, nil
}

func MigrateDB(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY NOT NULL,
			auth_uid VARCHAR(255) UNIQUE NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL,
			birth_date DATE NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate users: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			transaction_id INTEGER PRIMARY KEY NOT NULL,
			transaction_code VARCHAR(255) UNIQUE NOT NULL,
			user_id INTEGER NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE ON UPDATE CASCADE
		);
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (transaction_id) REFERENCES transactions (transaction_id) ON DELETE CASCADE ON UPDATE CASCADE,
			FOREIGN KEY (item_id) REFERENCES items (item_id) ON DELETE CASCADE ON UPDATE CASCADE
		);
	`); err != nil {
		return fmt.Errorf("Migrate(): failed to migrate transactions: %w", err)
	}

	return nil
}
