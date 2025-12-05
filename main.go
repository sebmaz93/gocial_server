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
	if dbURL == "" {
		log.Fatal("DB_URL variable must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	ENV := os.Getenv("ENV")
	if ENV == "" {
		log.Fatal("ENV variable must be set")
	}
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET variable must be set")
	}
	apiCfg := handlers.ApiConfig{
		FileserverHits: atomic.Int32{},
		DB:             dbQueries,
		ENV:            ENV,
		JWTSecret:      JWTSecret,
	}

	dir := http.Dir(rootPath)
	fileServer := http.FileServer(dir)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandleResetMetrics)
	mux.HandleFunc("GET /api/healthz", handlers.HandlerHealth)
	mux.HandleFunc("POST /api/users", apiCfg.HandleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.HandleLogin)
	mux.HandleFunc("POST /api/chirps", apiCfg.HandleCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.HandleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.HandleGetChirpByID)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandleRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.HandleRevokeToken)
	mux.HandleFunc("PUT /api/users", apiCfg.HandleUpdateUser)

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
