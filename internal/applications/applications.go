package applications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"test/internal/candidates"
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

type ApplicationWithCandidate struct {
	ID          uuid.UUID
	CandidateID string
	Role        string
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Candidate candidates.Candidate
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

func (r *Repository) GetApplications(
	role *string,
	status *string,
	limit int,
	offset int,
) ([]ApplicationWithCandidate, error) {
	rows, err := r.db.Query(
		`
			SELECT
				a.id,
				a.candidate_id,
				a.role,
				a.status,
				a.version,
				a.created_at,
				a.updated_at,
				c.id,
				c.first_name,
				c.last_name,
				c.email
			FROM applications a
			JOIN candidates c
				ON c.id = a.candidate_id
			WHERE
				($1::text IS NULL OR a.role = $1)
				AND
				($2::text IS NULL OR a.status = $2)
			ORDER BY a.created_at DESC
			LIMIT $3 OFFSET $4
		`,
		role,
		status,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applications := make([]ApplicationWithCandidate, 0)

	for rows.Next() {
		var app ApplicationWithCandidate
		var email string

		if err := rows.Scan(
			&app.ID,
			&app.CandidateID,
			&app.Role,
			&app.Status,
			&app.Version,
			&app.CreatedAt,
			&app.UpdatedAt,
			&app.Candidate.ID,
			&app.Candidate.FirstName,
			&app.Candidate.LastName,
			&email,
		); err != nil {
			return nil, err
		}
		

		app.Candidate.Email = mail.Address{
			Address: email,
		}
		
		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}


	return applications, nil
}

func (r *Repository) GetApplication(id string) (*Application, error) {
	var app Application

	err := r.db.QueryRow(
		`
			SELECT
				id,
				candidate_id,
				role,
				status,
				version,
				created_at,
				updated_at
			FROM applications
			WHERE id = $1;
		`,
		id,
	).Scan(
		&app.ID,
		&app.CandidateID,
		&app.Role,
		&app.Status,
		&app.Version,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // or custom NotFound error
		}
		return nil, err
	}

	return &app, nil
}

func (r *Repository) PromoteApplication(id string, status string) (*Application, error) {
	application, err := r.GetApplication(id)
	if err != nil {
		return nil, err
	}

	if application == nil {
		return application, err
	}

	if !canTransition(application.Status, status) {
		return nil, fmt.Errorf(
			"invalid status transition: %s → %s",
			application.Status,
			status,
		)
	} 

	var updated Application
	err = r.db.QueryRow(`
			UPDATE applications
			SET
				status = $1,
				version = version + 1,
				updated_at = now()
			WHERE id = $2
			RETURNING
				id,
				candidate_id,
				role,
				status,
				version,
				created_at,
				updated_at;
		`,
		status,
		application.ID,

	).Scan(
		&updated.ID,
		&updated.CandidateID,
		&updated.Role,
		&updated.Status,
		&updated.Version,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("application was updated concurrently")
		}
		return nil, err
	}

	return &updated, nil
}

var allowedTransitions = map[string][]string{
	"applied":   {"screening", "rejected"},
	"screening": {"interview", "rejected"},
	"interview": {"offer", "rejected"},
	"offer":     {"hired", "rejected"},
	"hired":     {},
	"rejected":  {},
}

func canTransition(from, to string) bool {
	nextStatuses, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	for _, status := range nextStatuses {
		if status == to {
			return true
		}
	}

	return false
}