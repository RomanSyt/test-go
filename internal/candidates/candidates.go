package candidates

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type Candidate struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     mail.Address
	CreatedAt time.Time
}

type CandidateData struct {
	FirstName string
	LastName  string
	Email     string
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

func (r *Repository) AddCandidate(
	firstName, lastName, email string,
) (*Candidate, error) {

	if firstName == "" || lastName == "" {
			return nil, errors.New("invalid name")
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil {
			return nil, fmt.Errorf("invalid email: %w", err)
	}

	c := Candidate{
			ID:        uuid.New(),
			FirstName: firstName,
			LastName:  lastName,
			Email:     *parsedEmail,
			CreatedAt: time.Now(),
	}

	_, err = r.db.Exec(`
			INSERT INTO candidates (id, first_name, last_name, email, created_at)
			VALUES ($1,$2,$3,$4,$5)
	`,
			c.ID,
			c.FirstName,
			c.LastName,
			c.Email.Address,
			c.CreatedAt,
	)

	if err != nil {
			return nil, err
	}

	return &c, nil
}