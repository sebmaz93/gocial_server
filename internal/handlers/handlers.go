package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/sebmaz93/gocial_server/internal/database"
	res "github.com/sebmaz93/gocial_server/internal/response"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	ENV            string
}

func HandlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func HandlerValidateChars(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	type resBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	data := reqBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		res.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(data.Body) > 140 {
		res.RespondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWordsMap := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(data.Body, badWordsMap)

	res.RespondWithJSON(w, http.StatusOK, resBody{
		CleanedBody: cleaned,
	})
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
