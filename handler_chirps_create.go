package main

import (
	"bootdev_chirpy/internal/database"
	"encoding/json"
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
	type requestBody struct {
		UserID uuid.UUID `json:"user_id"`
		Body   string    `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(reqBody.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	censureMsg := censureWords(reqBody.Body)

	chirpParam := database.CreateChirpParams{
		UserID: reqBody.UserID,
		Body:   censureMsg,
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParam)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	resBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 201, resBody)
}
