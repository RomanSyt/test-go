package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"test/internal/applications"
	"test/internal/candidates"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type CandidateData struct {
  FirstName string
  LastName string
  Email string
}

type ApplicationData struct {
  CandidateID string
	Role        string
}

type Server struct {
  candidatesManager *candidates.Manager
  applicationsManager *applications.Manager
}

func main() {
  _ = godotenv.Load()

  dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}

	defer db.Close()

  if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	log.Println("✅ database connected")

  // candidatesManager := candidates.NewManager()
  // applicationsManager := applications.NewManager()

  // s := Server{
  //   candidatesManager: candidatesManager,
  //   applicationsManager: applicationsManager,
  // }

  // mux := http.NewServeMux()

  // mux.HandleFunc("GET /candidates", s.getCandidates)
  // mux.HandleFunc("POST /candidates", s.addCandidate)
  // mux.HandleFunc("POST /get-candidates", s.getCandidate)
  // mux.HandleFunc("POST /applications", s.addApplication)
  // mux.HandleFunc("GET /applications", s.getApplications)

  // log.Fatal( http.ListenAndServe(":8080", mux))
}

func (s *Server) addCandidate(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r) {
		return
	}

  // limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

  var candidateData CandidateData

  err := decoder.Decode(&candidateData)

  if err != nil {
		slog.Error("error decoding addCandidate request body", "err", err)
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

  err = s.candidatesManager.AddCandidate(candidateData.FirstName, candidateData.LastName, candidateData.Email)

  if err != nil {
		http.Error(w, fmt.Sprintf("error adding candidate: %v\n", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getCandidate(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r) {
		return
	}

	// limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	var candidateData CandidateData

	err := decoder.Decode(&candidateData)
	if err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v\n", err), http.StatusBadRequest)
		return
	}

	candidate, err := s.candidatesManager.GetCandidateByName(candidateData.FirstName, candidateData.LastName)
	if err != nil {
		if errors.Is(err, candidates.ErrNoResultsFound) {
			http.Error(w, "no candidate found", http.StatusNotFound)
		} else {
			slog.Error("error retrieving candidate", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	converted := convertCandidateToCandidateData(candidate)

	marshalled, err := json.Marshal(converted)
	if err != nil {
		slog.Error("error marshalling getCandidate response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(marshalled)
	if err != nil {
		// headers are set by write call, best we can do is log an error
		slog.Error("error writing getCandidate response body", "err", err)
	}
}

func (s *Server) getCandidates(w http.ResponseWriter, r *http.Request) {
  marshalled, err := json.Marshal(s.candidatesManager.Candidates())
	if err != nil {
		slog.Error("error marshalling getCandidates response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

  w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(marshalled)
	if err != nil {
		// headers are set by write call, best we can do is log an error
		slog.Error("error writing getCandidates response body", "err", err)
	}
}

func (s *Server) addApplication(w http.ResponseWriter, r *http.Request) {
  if !validateContentType(w, r) {
		return
	}

  // limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

  var applicationData ApplicationData

  err := decoder.Decode(&applicationData)

  if err != nil {
		slog.Error("error decoding addApplication request body", "err", err)
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

  err = s.applicationsManager.AddApplication(applicationData.CandidateID, applicationData.Role)

  if err != nil {
		http.Error(w, fmt.Sprintf("error adding application: %v\n", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getApplications(w http.ResponseWriter, r *http.Request) {
  marshalled, err := json.Marshal(s.applicationsManager.Applications())
	if err != nil {
		slog.Error("error marshalling getApplications response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

  w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(marshalled)
	if err != nil {
		// headers are set by write call, best we can do is log an error
		slog.Error("error writing getApplications response body", "err", err)
	}
}

func convertCandidateToCandidateData(u *candidates.Candidate) *CandidateData {
	converted := CandidateData{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email.Address,
	}

	return &converted
}

func validateContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, fmt.Sprintf("unsupported Content-Type header %q", contentType), http.StatusUnsupportedMediaType)
		return false
  }
  return true
}