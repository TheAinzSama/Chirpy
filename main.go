package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/TheAinzSama/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secretKey      string
}

func main() {

	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to Open DB Connection: %v", err)
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	dbQueries := database.New(dbConn)
	apiCfg := &apiConfig{}
	apiCfg.dbQueries = dbQueries
	apiCfg.platform = platform
	apiCfg.secretKey = os.Getenv("SECRET_KEY")
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.findChirps)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUserCreate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserAuth)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
