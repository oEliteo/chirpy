package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type args struct {
		Body string `json:"body"`
	}

	type successResponse struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := args{}
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error\n")
		return
	}

	if len(params.Body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, "Chirp is too long\n")
		return
	}

	cfg.respondWithJSON(w, http.StatusOK, successResponse{Valid: true})
}
