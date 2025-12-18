package applications

import (
	"errors"
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

type Manager struct {
	applications []Application
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddApplication(candidateID string, role string)error {
	if candidateID == "" {
		return  errors.New("candidate id is required")
	}

	if role == "" {
		return  errors.New("role is required")
	}

	candidateUUID, err := uuid.Parse(candidateID)

	if err != nil {
		return  errors.New("candidate id is not valid")
	}

	app := Application{
		ID:          uuid.New(),
		CandidateID: candidateUUID,
		Role:        role,
		Status:      "applied",
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.applications = append(m.applications, app)

	return nil
}