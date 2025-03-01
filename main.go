package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/TheAinzSama/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	dbQueries := database.New(dbConn)
	apiCfg := &apiConfig{}
	apiCfg.dbQueries = dbQueries
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
