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
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Email      string    `json:"email"`
	Token      string    `json:"token"`
	IsChirpRed bool      `json:"is_chirpy_red"`
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
			ID:         user.ID,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
			Email:      user.Email,
			IsChirpRed: user.IsChirpyRed,
		},
	})
}

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(dbUser.ID, cfg.JWTSecret, defaultExpiresIn)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating JWT", err)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	err = cfg.DB.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating refresh JWT", err)
		return
	}

	res.RespondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:         dbUser.ID,
			CreatedAt:  dbUser.CreatedAt,
			UpdatedAt:  dbUser.UpdatedAt,
			Email:      dbUser.Email,
			IsChirpRed: dbUser.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *ApiConfig) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error reading refresh JWT", err)
		return
	}
	// TODO : get user from DB not from token
	dbToken, err := cfg.DB.GetRefreshToken(context.Background(), token)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error token not found or expired", err)
		return
	}
	i := time.Now().Compare(dbToken.ExpiresAt)
	if i >= 0 || dbToken.RevokedAt.Valid {
		res.RespondWithError(w, http.StatusUnauthorized, "Error token not found or expired", err)
		return
	}

	newToken, err := auth.MakeJWT(dbToken.UserID, cfg.JWTSecret, defaultExpiresIn)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error creating JWT", err)
		return
	}
	type response struct {
		Token string `json:"token"`
	}
	res.RespondWithJSON(w, http.StatusOK, response{
		Token: newToken,
	})
}

func (cfg *ApiConfig) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error reading refresh JWT", err)
		return
	}
	_, err = cfg.DB.GetRefreshToken(context.Background(), token)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error token not found or expired", err)
		return
	}

	err = cfg.DB.RevokeToken(context.Background(), token)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error revoking token", err)
		return
	}
	res.RespondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *ApiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error reading JWT", err)
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		res.RespondWithError(w, http.StatusUnauthorized, "Error validating JWT", err)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&params)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Hashing password error", err)
		return
	}

	updatedUser, err := cfg.DB.UpdateUser(context.Background(), database.UpdateUserParams{
		ID:             uuid,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error updating user info", err)
		return
	}
	type response struct {
		Email string `json:"email"`
	}

	res.RespondWithJSON(w, http.StatusOK, response{
		Email: updatedUser.Email,
	})
}

func (cfg *ApiConfig) HandlePolkaHook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.POLKA {
		res.RespondWithError(w, http.StatusUnauthorized, "Apikey error", err)
		return
	}

	type tRequestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	requestBody := tRequestBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&requestBody)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if requestBody.Event != "user.upgraded" {
		res.RespondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	parsedUserID, err := uuid.Parse(requestBody.Data.UserID)
	if err != nil {
		res.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	_, err = cfg.DB.UpgradeUserToRed(context.Background(), parsedUserID)
	if err != nil {
		res.RespondWithError(w, http.StatusNotFound, "Error updating user info", err)
		return
	}
	res.RespondWithJSON(w, http.StatusNoContent, nil)
}
