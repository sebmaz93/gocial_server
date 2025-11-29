package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sebmaz93/gocial_server/internal/database"
	"github.com/sebmaz93/gocial_server/internal/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading ENV file.")
	}
	const port = "8080"
	const rootPath = "."
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	apiCfg := handlers.ApiConfig{
		FileserverHits: atomic.Int32{},
		DB:             dbQueries,
	}

	dir := http.Dir(rootPath)
	fileServer := http.FileServer(dir)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandleResetMetrics)
	mux.HandleFunc("GET /api/healthz", handlers.HandlerHealth)
	mux.HandleFunc("POST /api/validate_chirp", handlers.HandlerValidateChars)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	slog.Info("Server listening on", "port", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start the server!")
	}
}
