package main

import (
	"bootdev_chirpy/internal/auth"
	"fmt"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Not able to extract the refresh token. Got error: %v", err))
		return
	}

	res, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token cannot be found.")
		return
	}
	if res.RevokedAt.Valid && res.RevokedAt.Time.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked.")
		return
	}
	if res.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired.")
		return
	}

	token, err := auth.MakeJWT(res.UserID, cfg.token, time.Duration(60*60)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot generate token. Error: %v", err))
		return
	}

	resBody := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	respondWithJSON(w, http.StatusAccepted, resBody)
}
