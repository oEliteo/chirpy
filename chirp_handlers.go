package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(params.Body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	params.Body = filterChirp(params.Body, bannedWordsMap)

	//Validation Succeeded Create DB Record...
	args := database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	}
	createdChirp, err := cfg.db.CreateChirp(r.Context(), args)
	if err != nil {
		log.Printf("Error creating chirp database record")
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	resp := chirp{
		ID:        createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserID:    createdChirp.UserID,
	}
	cfg.respondWithJSON(w, http.StatusCreated, resp)
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	responseChirps := make([]chirp, 0)
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error retrieving chirps from database.")
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	for _, item := range dbChirps {
		tmp := chirp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Body:      item.Body,
			UserID:    item.UserID,
		}
		responseChirps = append(responseChirps, tmp)
	}

	cfg.respondWithJSON(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Malformed uuid in request")
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error retrieving chirp from database")
		cfg.respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	responseChirp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	cfg.respondWithJSON(w, http.StatusOK, responseChirp)
}
