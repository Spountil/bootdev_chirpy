package main

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp ID not formatted correctly")
	}
	ctx := context.Background()
	chirp, err := cfg.dbQueries.GetChirp(ctx, chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp does not exist")
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
