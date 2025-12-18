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
	CandidateID string
	Role        string
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ApplicationData struct {
	CandidateID string
	Role        string
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

func (r *Repository) AddApplication(
	candidateID string,
	role string,
) (*Application, error) {
	a := Application{
		ID:        uuid.New(),
		CandidateID: candidateID,
		Role: role,
		Status: "applied",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO applications (id, candidate_id, role, status)
		VALUES ($1, $2, $3, 'applied')
	`,
		a.ID,
		a.CandidateID,
		a.Role,
	)

	return &a, err
}