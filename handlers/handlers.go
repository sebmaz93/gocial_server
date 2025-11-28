package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
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
		Error        string `json:"error"`
		Cleaned_body string `json:"cleaned_body"`
	}

	data := reqBody{}
	resData := resBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resData = resBody{
			Error: err.Error(),
		}
		dat, err := json.Marshal(resData)
		if err != nil {
			return
		}
		w.Write(dat)
		return
	}
	if len(data.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		resData = resBody{
			Error: "Chirp is too long",
		}
		dat, _ := json.Marshal(resData)
		w.Write(dat)
	}
	badWordsMap := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleaned := getCleanedBody(data.Body, badWordsMap)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resData = resBody{
		Cleaned_body: cleaned,
	}
	dat, _ := json.Marshal(resData)
	w.Write(dat)
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
