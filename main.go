package main

import (
	"context"
	"log"
	"os"
	"test/internal/candidates"
	"test/internal/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	if err := candidates.EnsureSchema(ctx, database); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ candidates table ready")


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