package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := apiConfig{}
	port := "8080"
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("GET /api/healthz", checkServerStatus)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", cfg.validateChirp)
	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func checkServerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK) + "\n"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK) + "\n"))
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type args struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := args{}
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	if len(params.Body) > 140 {
		cfg.respondWithError(w, fmt.Errorf("Chirp is too long"), http.StatusBadRequest)
		return
	}

	type successResponse struct {
		Valid bool `json:"valid"`
	}

	successResp := successResponse{
		Valid: true,
	}

	dat, err := json.Marshal(successResp)
	if err != nil {
		cfg.respondWithError(w, err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) respondWithError(w http.ResponseWriter, e error, status int) {
	type requestError struct {
		Msg string `json:"error"`
	}
	dat := make([]byte, 0)
	err := errors.New("")
	reqErr := requestError{}
	switch status {
	case http.StatusInternalServerError:
		reqErr.Msg = fmt.Sprintf("Internal Server Error: %v", status)
		dat, err = json.Marshal(reqErr)
	case http.StatusBadRequest:
		reqErr.Msg = fmt.Sprintf("Bad Request: %s", e)
		dat, err = json.Marshal(reqErr)
	default:
		reqErr.Msg = "An Unknown Error Has Occured"
		dat, err = json.Marshal(reqErr)
	}

	if err != nil {
		fmt.Printf("Error Marshalling reqErr: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("An unknown error has occured"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dat)
}
