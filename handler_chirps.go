package main

import (
	"bootdev_chirpy/internal/auth"
	"bootdev_chirpy/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"udpated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func censureWords(msg string) (censuredMsg string) {
	words := strings.Split(msg, " ")
	for i, word := range words {
		if _, exists := badWords[strings.ToLower(word)]; exists {
			words[i] = "****"
		}
	}
	censuredMsg = strings.Join(words, " ")
	return censuredMsg
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token missing in the Headers.")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Could not validate the token. got err: %v", err))
		return
	}

	type requestBody struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err = decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}
	censureMsg := censureWords(reqBody.Body)

	chirpParam := database.CreateChirpParams{
		UserID: userID,
		Body:   censureMsg,
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParam)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	resBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, resBody)
}

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
