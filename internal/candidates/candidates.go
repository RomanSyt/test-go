package candidates

import (
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

var ErrNoResultsFound error = errors.New("no results found")

type Candidate struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     mail.Address
	CreatedAt time.Time
}

type Manager struct {
	candidates []Candidate
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddCandidate(firstName string, lastName string, email string) error {
	if firstName == "" {
		return fmt.Errorf("invalid first name: %q", firstName)
	}

	if lastName == "" {
		return fmt.Errorf("invalid last name: %q", lastName)
	}

	existingCandidate, err := m.GetCandidateByName(firstName, lastName)
	if err != nil && !errors.Is(err, ErrNoResultsFound) {
		return fmt.Errorf("error checking if candidate is already present: %v", err)
	}

	if existingCandidate != nil {
		return errors.New("candidate with this name already exists")
	}

	parsedAddress, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email: %s", email)
	}

	newCandidate := Candidate{
		ID: uuid.New(),
		FirstName: firstName,
		LastName:  lastName,
		Email:     *parsedAddress,
		CreatedAt: time.Now(),
	}

	m.candidates = append(m.candidates, newCandidate)

	return nil
}

func (m *Manager) GetCandidateByName(first string, last string) (*Candidate, error) {
	for i, candidate := range m.candidates {
		if candidate.FirstName == first && candidate.LastName == last {
			result := m.candidates[i]
			return &result, nil
		}
	}
	return nil, ErrNoResultsFound
}
