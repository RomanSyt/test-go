package candidates

import (
	"context"
	"database/sql"
	"fmt"
)

const createCandidatesTableSQL = `
	CREATE TABLE IF NOT EXISTS candidates (
		id UUID PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT now()
	);
`

// EnsureSchema creates the candidates table if it does not exist.
// This function is safe to call multiple times.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	_, err := db.ExecContext(ctx, createCandidatesTableSQL)
	return err
}