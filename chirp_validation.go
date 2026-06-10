package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

var bannedWordsMap = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type args struct {
		Body string `json:"body"`
	}

	type successResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := args{}
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(params.Body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cfg.respondWithJSON(w, http.StatusOK, successResponse{CleanedBody: filterChirp(params.Body, bannedWordsMap)})
}

func filterChirp(body string, bannedWords map[string]struct{}) string {
	words := make([]string, 0)
	for word := range strings.FieldsSeq(body) {
		_, exists := bannedWords[strings.ToLower(word)]
		if exists {
			words = append(words, "****")
		} else {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}
