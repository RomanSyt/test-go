package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"test/internal/candidates"
	"test/internal/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Server struct {
  candidates *candidates.Repository
}

func main() {
  _ = godotenv.Load()

	cfg := db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	log.Println("✅ db connected")

	ctx := context.Background()

  s := Server{
    candidates: candidates.NewRepository(database),
  }

	if err := s.candidates.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	} else { 
    log.Println("✅ candidates table ready")
  }

  mux := http.NewServeMux()

  // mux.HandleFunc("GET /candidates", s.getCandidates)
  mux.HandleFunc("POST /candidates", s.addCandidate)

  log.Fatal( http.ListenAndServe(":8080", mux))
}

func (s *Server) addCandidate(w http.ResponseWriter, r *http.Request) {
  contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, fmt.Sprintf("unsupported Content-Type header %q", contentType), http.StatusUnsupportedMediaType)
		return
  }

  // limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

  var candidateData candidates.CandidateData

  err := decoder.Decode(&candidateData)
  if err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v\n", err), http.StatusBadRequest)
		return
	}

  candidate, err := s.candidates.AddCandidate(candidateData.FirstName, candidateData.LastName, candidateData.Email)
  if err != nil {
		http.Error(w, fmt.Sprintf("error adding candidate: %v\n", err), http.StatusBadRequest)
		return
	}

  
  marshalled, err := json.Marshal(candidate)
  if err != nil {
    slog.Error("error marshalling getCandidate response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
  _, err = w.Write(marshalled)
  if err != nil {
		slog.Error("error response body", "err", err)
	}
}