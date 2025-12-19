package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"test/internal/applications"
	applicationsevents "test/internal/applicationsEvents"
	"test/internal/candidates"
	"test/internal/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Server struct {
  candidates *candidates.Repository
  applications *applications.Repository
  applicationsEvents *applicationsevents.Repository
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
    applications: applications.NewRepository(database),
    applicationsEvents: applicationsevents.NewRepository(database),
  }

	if err := s.candidates.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	} else { 
    log.Println("✅ candidates table ready")
  }


	if err := s.applications.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	} else { 
    log.Println("✅ applications table ready")
  }


  if err := s.applicationsEvents.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	} else { 
    log.Println("✅ applications events table ready")
  }


  mux := http.NewServeMux()

  mux.HandleFunc("POST /candidates", s.addCandidate)
  mux.HandleFunc("POST /applications", s.addApplication)
  mux.HandleFunc("GET /applications", s.getApplications)
  mux.HandleFunc("GET /applications/{id}", s.getApplication)
  mux.HandleFunc("POST /applications/{id}/transition", s.addApplicationEvent)

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
    slog.Error("error marshalling response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
  _, err = w.Write(marshalled)
  if err != nil {
		slog.Error("error response body", "err", err)
	}
}

func (s *Server) addApplication(w http.ResponseWriter, r *http.Request) {
  contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, fmt.Sprintf("unsupported Content-Type header %q", contentType), http.StatusUnsupportedMediaType)
		return
  }

  // limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

  var applicationData applications.ApplicationData

  err := decoder.Decode(&applicationData)
  if err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v\n", err), http.StatusBadRequest)
		return
	}

  application, err := s.applications.AddApplication(applicationData.CandidateID,applicationData.Role)
  if err != nil {
		http.Error(w, fmt.Sprintf("error adding candidate: %v\n", err), http.StatusBadRequest)
		return
	}

  marshalled, err := json.Marshal(application)
  if err != nil {
    slog.Error("error marshalling response", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
  _, err = w.Write(marshalled)
  if err != nil {
		slog.Error("error response body", "err", err)
	}
}

func (s *Server) getApplications(w http.ResponseWriter, r *http.Request) {
  params := r.URL.Query()

  var role *string
  if v := params.Get("user"); v != "" {
      role = &v
  }

  var status *string
  if v := params.Get("status"); v != "" {
      status = &v
  }

  limit := 20
  if v := params.Get("limit"); v != "" {
      if parsed, err := strconv.Atoi(v); err == nil {
          limit = parsed
      }
  }

  offset := 0
  if v := params.Get("offset"); v != "" {
      if parsed, err := strconv.Atoi(v); err == nil {
          offset = parsed
      }
  }

  applications, err := s.applications.GetApplications(role, status, limit, offset)
  if err != nil {
    slog.Error("error GetApplications", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

  w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(applications); err != nil {
		slog.Error("json encode error", "err", err)
	}
}


func (s *Server) getApplication(w http.ResponseWriter, r *http.Request) {
  id := r.PathValue("id")

  application, err := s.applications.GetApplication(id)
  if err != nil {
    slog.Error("error GetApplication", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

  w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(application); err != nil {
		slog.Error("json encode error", "err", err)
	}
}

func (s *Server) addApplicationEvent(w http.ResponseWriter, r *http.Request) {
  contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, fmt.Sprintf("unsupported Content-Type header %q", contentType), http.StatusUnsupportedMediaType)
		return
  }

  // limit to 1MB
	requestBody := http.MaxBytesReader(w, r.Body, 1048576)
  
	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()
  
  var bodyData applicationsevents.ApplicationEventBody
  
  err := decoder.Decode(&bodyData)
  if err != nil {
    http.Error(w, fmt.Sprintf("error decoding request body: %v\n", err), http.StatusBadRequest)
		return
	}

  id := r.PathValue("id")
  updated, err := s.applications.PromoteApplication(id, bodyData.ToStatus)

  if err != nil {
    slog.Error("error addApplicationEvent", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

  payload, err := json.Marshal(bodyData)
  if err != nil {
    http.Error(
      w,
      fmt.Sprintf("error encoding payload: %v\n", err),
      http.StatusInternalServerError,
    )
    return
  }

  s.applicationsEvents.AddApplicationEvent(&applicationsevents.ApplicationEventData{
    ApplicationID: id,
    Type: "update",
    Payload: payload,
  })

  w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updated); err != nil {
		slog.Error("json encode error", "err", err)
	}
}