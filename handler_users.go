package main

import (
	"bootdev_chirpy/internal/auth"
	"bootdev_chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
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

	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to hash password")
		return
	}

	args := database.CreateUserParams{
		Email: reqBody.Email,
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), args)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to create user. Error %v", err))
		return
	}

	resBody := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusCreated, resBody)
}
