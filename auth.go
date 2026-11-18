package main

import (
	"bootdev_chirpy/internal/auth"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) authorization(w http.ResponseWriter, headers http.Header) uuid.UUID {
	token, err := auth.GetBearerToken(headers)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Not able to extract the token. Got error: %v", err))
		return uuid.Nil
	}

	userID, err := auth.ValidateJWT(token, cfg.token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Not able to ID from token. Got error: %v", err))
		return uuid.Nil
	}

	return userID
}
