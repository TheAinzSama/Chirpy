package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheAinzSama/Chirpy/internal/auth"
	"github.com/TheAinzSama/Chirpy/internal/database"
)

type userInfo struct {
	ID            string `json:"id"`
	Created       string `json:"created_at"`
	Updated       string `json:"updated_at"`
	Email         string `json:"email"`
	Token         string `json:"token"`
	Refresh_token string `json:"refresh_token"`
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
	token, err := auth.MakeJWT(user.ID, cfg.secretKey, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Token", err)
		return
	}
	expiresAt := time.Now().AddDate(0, 0, 60)
	refresT, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Refresh Token1", err)
		return
	}
	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refresT,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	refreshToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Refresh Token2", err)
		return
	}
	respondWithJSON(w, http.StatusOK, userInfo{
		ID:            user.ID.String(),
		Created:       user.CreatedAt.String(),
		Updated:       user.UpdatedAt.String(),
		Email:         user.Email,
		Token:         token,
		Refresh_token: refreshToken.Token,
	})
}
func (cfg *apiConfig) handlerCheckRefreshToken(w http.ResponseWriter, r *http.Request) {
	reqAuthheader := r.Header.Get("Authorization")
	if reqAuthheader == "" {
		respondWithError(w, http.StatusUnauthorized, "Empty Refresh Token", nil)
		return
	}
	refreshToken, err := cfg.dbQueries.SelectRefreshToken(r.Context(), reqAuthheader[7:])
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", nil)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}
	userID, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get ID From Refresh Token", err)
		return
	}
	token, err := auth.MakeJWT(userID.ID, cfg.secretKey, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create New Access Token", err)
		return
	}
	respondWithJSON(w, http.StatusOK, userInfo{
		Token: token,
	})
}
func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	reqAuthheader := r.Header.Get("Authorization")
	if reqAuthheader == "" {
		respondWithError(w, http.StatusUnauthorized, "Empty Refresh Token", nil)
		return
	}
	err := cfg.dbQueries.RevokeRefreshToken(r.Context(), reqAuthheader[7:])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	reqAuthheader := r.Header.Get("Authorization")
	if reqAuthheader == "" {
		respondWithError(w, http.StatusUnauthorized, "Empty Refresh Token", nil)
		return
	}
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
	tokenUserID, err := auth.ValidateJWT(reqAuthheader[7:], cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get UserID", err)
		return
	}
	hpass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't Hash the password ", err)
		return
	}
	createParams := database.UpdateUserInfoParams{
		ID:             tokenUserID,
		HashedPassword: hpass,
	}
	dberr := cfg.dbQueries.UpdateUserInfo(r.Context(), createParams)
	if dberr != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't Update User", err)
		return
	}
	respondWithJSON(w, http.StatusOK, userInfo{
		Email: params.Email,
	})
}
