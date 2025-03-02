package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	type userInfo struct {
		ID      string `json:"id"`
		Created string `json:"created_at"`
		Updated string `json:"updated_at"`
		Email   string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := userInfo{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't Create User", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, userInfo{
		ID:      user.ID.String(),
		Created: user.CreatedAt.String(),
		Updated: user.UpdatedAt.String(),
		Email:   user.Email,
	})
}
