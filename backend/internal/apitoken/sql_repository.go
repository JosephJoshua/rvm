package apitoken

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/apitoken/domain"
	"github.com/jmoiron/sqlx"
)

type apiToken struct {
	APITokenID string       `db:"api_token_id"`
	ExpiringAt sql.NullTime `db:"expiring_at"`
	CreatedAt  sql.NullTime `db:"created_at"`
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) GetTokenByID(tokenID string) (*domain.APIToken, error) {
	var rawToken apiToken
	if err := r.db.Get(&rawToken, `
		SELECT
			api_token_id, expiring_at, created_at
		FROM
			api_tokens
		WHERE
			api_token_id = ?
		`, tokenID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}

		return nil, fmt.Errorf("GetTokenByID(): failed to execute query: %w", err)
	}

	var expiringAt *time.Time
	if rawToken.ExpiringAt.Valid {
		expiringAt = &rawToken.ExpiringAt.Time
	}

	token := domain.NewAPIToken(rawToken.APITokenID, expiringAt, rawToken.CreatedAt.Time)
	return token, nil
}
