package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	res "github.com/sebmaz93/gocial_server/internal/response"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type responseBody struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	user, err := cfg.DB.CreateUser(context.Background(), params.Email)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	res.RespondWithJSON(w, http.StatusCreated, responseBody{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})

}
