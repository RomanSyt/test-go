package applications

import "time"

type Application struct {
	ID          string
	CandidateID string
	Role        string
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}