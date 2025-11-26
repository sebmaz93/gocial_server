package handlers

import (
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := cfg.FileserverHits.Load()
	body := fmt.Sprintf("Hits: %d\n", hits)

	w.Write([]byte(body))
}

func (cfg *ApiConfig) HandleResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
