package main

import (
	"bootdev_chirpy/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	authKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Unable to extract the API Key. Got error: %v", err))
		return
	}

	if authKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	type RequestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := RequestBody{}
	err = decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode JSON")
		return
	}

	if reqBody.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}

	err = cfg.dbQueries.AddChirpyRed(r.Context(), reqBody.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Unable to update user. Got error: %v", err))
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
