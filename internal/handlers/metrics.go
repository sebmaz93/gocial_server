package handlers

import (
	"context"
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := cfg.FileserverHits.Load()
	body := fmt.Sprintf(`
	<html>
		<body>
    		<h1>Welcome, Chirpy Admin</h1>
      		<p>Chirpy has been visited %d times!</p>
        </body>
    </html>
	`, hits)

	w.Write([]byte(body))
}

func (cfg *ApiConfig) HandleResetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.ENV != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.FileserverHits.Store(0)
	err := cfg.DB.DeleteAllUsers(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error resetting"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
