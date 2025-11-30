package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
	}

	type responseBody struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Email      string    `json:"email"`
	}

	reqBody := requestBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&reqBody)
	if err != nil {
		fmt.Errorf("error decoding")
		return
	}

	user, err := cfg.DB.CreateUser(context.Background(), reqBody.Email)
	if err != nil {
		fmt.Errorf("error creating user")
	}

	resBody := responseBody{
		Id:         user.ID,
		Created_at: user.CreatedAt.Time,
		Updated_at: user.UpdatedAt.Time,
		Email:      user.Email,
	}
	dat, _ := json.Marshal(resBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)
}
