package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/TheAinzSama/Chirpy/internal/auth"
	"github.com/TheAinzSama/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirp struct {
	Id         string `json:"id"`
	Body       string `json:"body"`
	User_id    string `json:"user_id"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
func (apiCfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	type tempVals struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := tempVals{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't decode parameters", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}
	tokenUserID, err := auth.ValidateJWT(token, apiCfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get UserID", err)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := findAndReplace(params.Body)
	createParams := database.CreateChirpParams{
		Body:   cleanBody,
		UserID: tokenUserID,
	}
	user, err := apiCfg.dbQueries.CreateChirp(r.Context(), createParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create Chirp", nil)
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp{
		Id:         user.ID.String(),
		Created_at: user.CreatedAt.String(),
		Updated_at: user.UpdatedAt.String(),
		Body:       user.Body,
		User_id:    user.UserID.String(),
	})
}
func (apiCfg *apiConfig) handlerAllChirps(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Failed to parse User's ID", nil)
			return
		}
		chirps, err = apiCfg.dbQueries.SelectManyChirp(r.Context(), userUUID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed fetch User's Chirps", nil)
			return
		}
	} else {
		var err error
		chirps, err = apiCfg.dbQueries.SelectChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed fetch Chirps", nil)
			return
		}
	}

	var chirpArray []chirp
	for _, achirp := range chirps {
		newChirp := chirp{
			Id:         achirp.ID.String(),
			Created_at: achirp.CreatedAt.String(),
			Updated_at: achirp.UpdatedAt.String(),
			Body:       achirp.Body,
			User_id:    achirp.UserID.String(),
		}
		chirpArray = append(chirpArray, newChirp)
	}
	respondWithJSON(w, http.StatusOK, chirpArray)
}
func (apiCfg *apiConfig) findChirps(w http.ResponseWriter, r *http.Request) {
	achirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "NO Chirp ID was provided", nil)
		return
	}
	achirps, err := apiCfg.dbQueries.SelectChirp(r.Context(), achirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to find Chirp", nil)
		return
	}
	respondWithJSON(w, http.StatusOK, chirp{
		Id:         achirps.ID.String(),
		Created_at: achirps.CreatedAt.String(),
		Updated_at: achirps.UpdatedAt.String(),
		Body:       achirps.Body,
		User_id:    achirps.UserID.String(),
	})
}

func findAndReplace(body string) string {
	var badWords = []string{"kerfuffle", "sharbert", "fornax"}
	bodyList := strings.Split(body, " ")
	for i, bodyWord := range bodyList {
		for _, word := range badWords {
			if word == strings.ToLower(bodyWord) {
				bodyList[i] = "****"
			}
		}
	}
	return strings.Join(bodyList, " ")
}
func (apiCfg *apiConfig) handlerDeleteChirps(w http.ResponseWriter, r *http.Request) {
	reqAuthheader := r.Header.Get("Authorization")
	if reqAuthheader == "" {
		respondWithError(w, http.StatusUnauthorized, "Empty Refresh Token", nil)
		return
	}
	userID, err := auth.ValidateJWT(reqAuthheader[7:], apiCfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Couldn't get UserID for Chyrp", err)
		return
	}

	achirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NO Chirp ID was provided", nil)
		return
	}
	bchirps, err := apiCfg.dbQueries.SelectChirp(r.Context(), achirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find Chirp", nil)
		return
	}
	if userID != bchirps.UserID {
		respondWithError(w, http.StatusForbidden, "You have no rights to delete this Chirp", err)
		return
	}
	dBerr := apiCfg.dbQueries.DeleteChirp(r.Context(), achirpID)
	if dBerr != nil {
		respondWithError(w, http.StatusNotFound, "NO Chirp ID was provided", nil)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
