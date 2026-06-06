package main

import (
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
	mux.HandleFunc("/metrics", cfg.metricsHandler)
	mux.HandleFunc("/healthz", checkServerStatus)
	mux.HandleFunc("/reset", cfg.resetHandler)
	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func checkServerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
