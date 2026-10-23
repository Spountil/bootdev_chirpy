package main

import (
	"context"
	"net/http"
)

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	chirps, err := cfg.dbQueries.GetAllChrips(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	resChirps := []Chirp{}
	for _, chirp := range chirps {
		resChirps = append(resChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, resChirps)
}
