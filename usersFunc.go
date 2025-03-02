package main

import (
	"encoding/json"
	"net/http"

	"github.com/TheAinzSama/Chirpy/internal/auth"
	"github.com/TheAinzSama/Chirpy/internal/database"
)

type userInfo struct {
	ID      string `json:"id"`
	Created string `json:"created_at"`
	Updated string `json:"updated_at"`
	Email   string `json:"email"`
}
type userAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := userAuth{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Password == "" {
		respondWithError(w, http.StatusInternalServerError, "Empty Password", err)
		return
	}
	hpass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't Hash the password ", err)
		return
	}
	createParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hpass,
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), createParams)
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
func (cfg *apiConfig) handlerUserAuth(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userAuth{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Password == "" {
		respondWithError(w, http.StatusInternalServerError, "Empty Password", err)
		return
	}
	user, err := cfg.dbQueries.SelectUserInfo(r.Context(), params.Email)
	if auth.CheckPasswordHash(params.Password, user.HashedPassword) != nil {
		respondWithError(w, http.StatusUnauthorized, "You Have no place here!!!", err)
		return
	}
	respondWithJSON(w, http.StatusOK, userInfo{
		ID:      user.ID.String(),
		Created: user.CreatedAt.String(),
		Updated: user.UpdatedAt.String(),
		Email:   user.Email,
	})
}
