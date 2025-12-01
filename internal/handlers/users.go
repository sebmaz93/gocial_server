package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sebmaz93/gocial_server/internal/auth"
	"github.com/sebmaz93/gocial_server/internal/database"
	res "github.com/sebmaz93/gocial_server/internal/response"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

const defaultExpiresIn = time.Hour * 1

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}
	user, err := cfg.DB.CreateUser(context.Background(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})
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

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn *int   `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	dbUser, err := cfg.DB.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	ok, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !ok {
		res.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}
	expireTime := defaultExpiresIn
	if params.ExpiresIn != nil && time.Duration(*params.ExpiresIn) < defaultExpiresIn {
		if *params.ExpiresIn > 0 {
			expireTime = time.Duration(*params.ExpiresIn)
		}
	}
	token, err := auth.MakeJWT(dbUser.ID, cfg.JWTSecret, expireTime)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating JWT", err)
		return
	}

	res.RespondWithJSON(w, http.StatusOK, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	})
}
