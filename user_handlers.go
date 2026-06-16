package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
		ID:          dbUsr.ID,
		CreatedAt:   dbUsr.CreatedAt,
		UpdatedAt:   dbUsr.UpdatedAt,
		Email:       dbUsr.Email,
		IsChirpyRed: dbUsr.IsChirpyRed,
	}

	cfg.respondWithJSON(w, http.StatusCreated, usr)
}

func (cfg *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type loginUser struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %v", err)
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	dbUsr, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUsr.HashedPassword)
	if err != nil || !match {
		log.Printf("Failed to check password against hash: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	accessExpiry := time.Hour

	accessToken, err := auth.MakeJWT(dbUsr.ID, cfg.jwtSecret, accessExpiry)
	if err != nil {
		log.Printf("error creating jwt with given user")
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	refreshExpiry := time.Now().AddDate(0, 0, 60)

	dbRefreshTokenArgs := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    dbUsr.ID,
		ExpiresAt: refreshExpiry,
		RevokedAt: sql.NullTime{},
	}
	dbRefreshToken, err := cfg.db.CreateRefreshToken(r.Context(), dbRefreshTokenArgs)
	if err != nil {
		log.Printf("error creating refreshToken entry in database: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUsr := loginUser{
		ID:           dbUsr.ID,
		CreatedAt:    dbUsr.CreatedAt,
		UpdatedAt:    dbUsr.UpdatedAt,
		Email:        dbUsr.Email,
		Token:        accessToken,
		RefreshToken: dbRefreshToken.Token,
		IsChirpyRed:  dbUsr.IsChirpyRed,
	}
	cfg.respondWithJSON(w, http.StatusOK, respUsr)
}

func (cfg *apiConfig) userRefreshHandler(w http.ResponseWriter, r *http.Request) {
	type responseToken struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting bearer token from header: %v", err)
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	dbUsr, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("error getting user from provided refresh token: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	newToken, err := auth.MakeJWT(dbUsr.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		log.Printf("error creating new jwt for user: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respToken := responseToken{
		Token: newToken,
	}

	cfg.respondWithJSON(w, http.StatusOK, respToken)
}

func (cfg *apiConfig) userRevokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting bearer token from given header: %v", err)
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	revokeArgs := database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		Token:     token,
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), revokeArgs)
	if err != nil {
		log.Printf("error revoking refresh token: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type responseUser struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

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

	newHashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("error hashing given password: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	updateUserArgs := database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: newHashedPassword,
		UpdatedAt:      time.Now().UTC(),
	}

	dbUsr, err := cfg.db.UpdateUser(r.Context(), updateUserArgs)
	if err != nil {
		log.Printf("error updating user: %v", err)
		cfg.respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUser := responseUser{
		ID:          dbUsr.ID,
		CreatedAt:   dbUsr.CreatedAt,
		UpdatedAt:   dbUsr.UpdatedAt,
		Email:       dbUsr.Email,
		IsChirpyRed: dbUsr.IsChirpyRed,
	}

	cfg.respondWithJSON(w, http.StatusOK, respUser)
}

func (cfg *apiConfig) userUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("error getting api key from header: %v", err)
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if key != cfg.polkaKey {
		log.Printf("api key mismatch")
		cfg.respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %v", err)
		cfg.respondWithError(w, http.StatusBadRequest, "bad request")
		return
	}

	if params.Event != "user.upgraded" {
		log.Printf("polka event was not user.upgrade")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.SetIsChirpyRedByUserID(r.Context(), params.Data.UserID)
	if err != nil {
		log.Printf("error user not found: %v", err)
		cfg.respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
