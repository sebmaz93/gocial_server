package handlers

import (
	"net/http"
	"sync/atomic"

	"github.com/sebmaz93/gocial_server/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	ENV            string
	JWTSecret      string
	POLKA          string
}

func HandlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
