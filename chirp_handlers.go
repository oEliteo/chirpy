package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting bearer token from header: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("error validating token from header: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	if len(params.Body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, "chirp is too long")
		return
	}

	params.Body = filterChirp(params.Body, bannedWordsMap)

	args := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
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

	queryString := r.URL.Query().Get("author_id")

	if queryString == "" {
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
	} else {
		authorID, err := uuid.Parse(queryString)

		if err != nil {
			log.Printf("error parsing uuid from string: %v", err)
			cfg.respondWithError(w, http.StatusBadRequest, "bad request")
			return
		}

		dbChirps, err := cfg.db.GetChirpsByAuthorID(r.Context(), authorID)
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
	}

	sortString := r.URL.Query().Get("sort")

	if sortString == "desc" {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt)
		})
	} else {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.Before(responseChirps[j].CreatedAt)
		})
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

func (cfg *apiConfig) deleteChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Malformed uuid in request")
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("error getting chirp from database: %v", err)
		cfg.respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting bearer token from header: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("error validating jwt: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if dbChirp.UserID != userID {
		log.Printf("user is not the owner of the provided chirp")
		cfg.respondWithError(w, http.StatusForbidden, "forbidden")
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		log.Printf("error deleting chirp from database: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
