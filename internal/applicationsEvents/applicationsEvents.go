package applicationsevents

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ApplicationEvent struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	Type          string
	Payload       []byte // JSON
	CreatedAt     time.Time
}

type ApplicationEventData struct {
	ApplicationID string
	Type string
	Payload []byte
}

type ApplicationEventBody struct {
	ToStatus string
	Reason string
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
		CREATE TABLE IF NOT EXISTS application_events (
    	id UUID PRIMARY KEY,
    	application_id UUID NOT NULL,
    	type TEXT NOT NULL,
    	payload JSONB,
    	created_at TIMESTAMP NOT NULL DEFAULT now(),

    	CONSTRAINT fk_application
    	    FOREIGN KEY (application_id)
    	    REFERENCES applications(id)
			);
	`)

	return err
}

func (r *Repository) Add(event *ApplicationEventData) error {
	_, err := r.db.Exec(`
			INSERT INTO application_events (
				id,
				application_id,
				type,
				payload,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5);
		`,
		uuid.NewString(),
		event.ApplicationID,
		event.Type,
		event.Payload,
		time.Now(),
	)

	return err
}
