package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) newUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %v", err)
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	userArgs := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	}

	dbUsr, err := cfg.db.CreateUser(r.Context(), userArgs)
	if err != nil {
		log.Printf("Error creating new user: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	usr := User{
		ID:        dbUsr.ID,
		CreatedAt: dbUsr.CreatedAt,
		UpdatedAt: dbUsr.UpdatedAt,
		Email:     dbUsr.Email,
	}

	cfg.respondWithJSON(w, http.StatusCreated, usr)
}

func (cfg *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	dbUsr, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUsr.HashedPassword)
	if err != nil {
		log.Printf("Failed to check password against hash: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if match {
		respUsr := User{
			ID:        dbUsr.ID,
			CreatedAt: dbUsr.CreatedAt,
			UpdatedAt: dbUsr.UpdatedAt,
			Email:     dbUsr.Email,
		}
		cfg.respondWithJSON(w, http.StatusOK, respUsr)
	} else {
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
	}
}
