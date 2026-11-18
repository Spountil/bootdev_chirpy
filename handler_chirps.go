package main

import (
	"bootdev_chirpy/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
	userID := cfg.authorization(w, r.Header)
	if userID == uuid.Nil {
		return
	}

	type requestBody struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)
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
	paramAuthorID := r.URL.Query().Get("author_id")
	paramSort := r.URL.Query().Get("sort")
	ctx := context.Background()

	chirps := []database.Chirp{}
	var err error
	if paramAuthorID != "" {
		paramAuthorIDParsed, err := uuid.Parse(paramAuthorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		chirps, err = cfg.dbQueries.GetChripsByAuthor(ctx, paramAuthorIDParsed)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
	} else {
		chirps, err = cfg.dbQueries.GetAllChrips(ctx)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
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

	if paramSort == "desc" {
		sort.Slice(resChirps, func(i, j int) bool {
			return resChirps[i].CreatedAt.After(resChirps[j].CreatedAt)
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

func (cfg *apiConfig) handlerDelChirps(w http.ResponseWriter, r *http.Request) {
	userID := cfg.authorization(w, r.Header)
	if userID == uuid.Nil {
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp ID not formatted correctly")
		return
	}

	chirpInfo, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Can't find Chirp. Got err: %v", err))
		return
	}

	if userID != chirpInfo.UserID {
		respondWithError(w, http.StatusForbidden, "")
		return
	}

	err = cfg.dbQueries.DelChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Could not find Chirp to delete. Got err: %v", err))
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
