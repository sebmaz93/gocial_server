package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebmaz93/gocial_server/internal/auth"
	"github.com/sebmaz93/gocial_server/internal/database"
	res "github.com/sebmaz93/gocial_server/internal/response"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "error getting JWT", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "error validating token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	requestBody := reqBody{}
	defer r.Body.Close()
	err = decoder.Decode(&requestBody)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleaned, err := validateChirp(requestBody.Body)
	if err != nil {
		res.RespondWithError(w, http.StatusBadRequest, "error sanitizng chirp", err)
		return
	}

	chirp, err := cfg.DB.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userId,
	})
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
	}
	res.RespondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *ApiConfig) HandleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")
	sortOrder := "asc"
	if sortQuery == "desc" {
		sortOrder = "desc"
	}
	var dbChirps []database.Chirp
	var err error
	if authorID != "" {
		parsedID, err := uuid.Parse(authorID)
		if err != nil {
			res.RespondWithError(w, http.StatusInternalServerError, "malformed user id", err)
			return
		}
		dbChirps, err = cfg.DB.GetChirpsByAuthorID(context.Background(), database.GetChirpsByAuthorIDParams{
			UserID:  parsedID,
			Column2: sortOrder,
		})
		if err != nil {
			res.RespondWithError(w, http.StatusInternalServerError, "Error fetching chirps", err)
			return
		}
	} else {
		dbChirps, err = cfg.DB.GetAllChirps(context.Background(), sortOrder)
		if err != nil {
			res.RespondWithError(w, http.StatusInternalServerError, "Error fetching chirps", err)
			return
		}
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}
	res.RespondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *ApiConfig) HandleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		res.RespondWithError(w, http.StatusBadRequest, "Chirp ID missing", nil)
		return
	}

	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		res.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.DB.GetChirpByID(context.Background(), parsedChirpID)
	if err != nil {
		res.RespondWithError(w, http.StatusNotFound, "Error fetching Chirp", err)
		return
	}

	res.RespondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *ApiConfig) HandleDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		res.RespondWithError(w, http.StatusBadRequest, "Chirp ID missing", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error reading JWT", err)
		return
	}

	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		res.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error validating JWT", err)
		return
	}

	chirp, err := cfg.DB.GetChirpByID(context.Background(), parsedChirpID)
	if err != nil {
		res.RespondWithError(w, http.StatusNotFound, "Error fetching Chirp", err)
		return
	}

	if chirp.UserID != userID {
		res.RespondWithError(w, http.StatusForbidden, "You can't delete this Chirp", err)
		return
	}

	err = cfg.DB.DeleteChirpById(context.Background(), chirp.ID)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error deleting Chirp", err)
		return
	}

	res.RespondWithJSON(w, http.StatusNoContent, nil)
}
