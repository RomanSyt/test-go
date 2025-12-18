package candidates

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS candidates (
			id UUID PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	`)
	return err
}