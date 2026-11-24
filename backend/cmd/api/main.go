package main

import (
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"calorietrack/backend/internal/application"
	"calorietrack/backend/internal/infrastructure/postgres"
	httptransport "calorietrack/backend/internal/transport/http"
)

func main() {
	addr := env("CALORIETRACK_ADDR", ":8080")
	databaseURL := env("CALORIETRACK_DATABASE_URL", "postgres://calorietrack:calorietrack_local@localhost:5432/calorietrack?sslmode=disable")
	allowedOrigin := env("CALORIETRACK_ALLOWED_ORIGIN", "http://localhost:5173")

	db, err := postgres.Open(databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	repository := postgres.NewRepository(db)
	service := application.NewService(repository)
	handler := httptransport.NewHandler(service, allowedOrigin)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("CalorieTrack API listening on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
