package main

import (
	"bootdev_chirpy/internal/auth"
	"bootdev_chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode JSON")
		return
	}

	loginInfo, err := cfg.dbQueries.Login(r.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to find user. Error %v", err))
		return
	}

	isSame, err := auth.CheckPasswordHash(reqBody.Password, loginInfo.HashedPassword.String)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong. Error %v", err))
		return
	}

	expirationInDuration := time.Duration(60*60) * time.Second

	token, err := auth.MakeJWT(loginInfo.ID, cfg.token, expirationInDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong. Error %v", err))
		return
	}

	refreshToken := auth.MakeRefreshToken()
	args := database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: loginInfo.ID,
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), args)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong. Error %v", err))
		return
	}

	resBody := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}{
		ID:           loginInfo.ID,
		CreatedAt:    loginInfo.CreatedAt,
		UpdatedAt:    loginInfo.UpdatedAt,
		Email:        loginInfo.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  loginInfo.IsChirpyRed.Bool,
	}

	if isSame {
		respondWithJSON(w, http.StatusOK, resBody)
	} else {
		respondWithError(w, http.StatusUnauthorized, "")
	}
}
