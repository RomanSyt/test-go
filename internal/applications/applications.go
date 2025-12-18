package applications

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID          uuid.UUID
	CandidateID uuid.UUID
	Role        string
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EnsureSchema(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}

	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS applications (
		  id UUID PRIMARY KEY,
		  candidate_id UUID NOT NULL REFERENCES candidates(id),
		  role TEXT NOT NULL,
		  status TEXT NOT NULL,
		  version INT NOT NULL DEFAULT 1,
		  created_at TIMESTAMP NOT NULL DEFAULT now(),
		  updated_at TIMESTAMP NOT NULL DEFAULT now()
		);
	`)
	return err
}