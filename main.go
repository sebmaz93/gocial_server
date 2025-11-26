package main

import (
	"log"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/sebmaz93/gocial_server/handlers"
)

func main() {
	const port = "8080"
	const rootPath = "."

	apiCfg := handlers.ApiConfig{
		FileserverHits: atomic.Int32{},
	}

	dir := http.Dir(rootPath)
	fileServer := http.FileServer(dir)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("/metrics", apiCfg.HandleMetrics)
	mux.HandleFunc("/reset", apiCfg.HandleResetMetrics)
	mux.HandleFunc("/healthz", handlers.HandlerHealth)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	slog.Info("Server listening on", "port", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start the server!")
	}
}
