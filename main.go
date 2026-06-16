package main

import (
	"chirpy/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Could not open database connection: %s", err)
	}

	dbQueries := database.New(db)

	cfg := apiConfig{
		db:        dbQueries,
		platform:  platform,
		jwtSecret: secret,
		polkaKey:  polkaKey,
	}
	port := "8080"
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("GET /api/healthz", cfg.checkServerStatus)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", cfg.postChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)
	mux.HandleFunc("POST /api/users", cfg.newUserHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpByIDHandler)
	mux.HandleFunc("POST /api/login", cfg.userLoginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.userRefreshHandler)
	mux.HandleFunc("POST /api/revoke", cfg.userRevokeHandler)
	mux.HandleFunc("PUT /api/users", cfg.userUpdateHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpByIDHandler)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.userUpgradeHandler)
	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) checkServerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK) + "\n"))
}
