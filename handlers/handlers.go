package handlers

import (
	"encoding/json"
	"net/http"
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
		Error string `json:"error"`
		Valid bool   `json:"valid"`
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resData = resBody{
		Valid: true,
	}
	dat, _ := json.Marshal(resData)
	w.Write(dat)
}
