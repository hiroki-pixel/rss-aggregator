package main

import (
	"encoding/json"
	"net/http"
	"rss-aggregator/internal/database"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateFeedFollow(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
) {
	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	params := parameters{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
		return
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    params.FeedID,
	}

	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), feedFollowParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create feed follow")
		return
	}

	respondWithJson(w, http.StatusCreated, databaseFeedFollowToFeedFollow(feedFollow))
}

func (apiCfg *apiConfig) handlerGetFeedFollows(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
) {
	feedFollows, err := apiCfg.DB.GetFeedFollowsForUsers(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get feed follows")
		return
	}

	respondWithJson(w, http.StatusOK, databaseFeedFollowsToFeedFollws(feedFollows))

}

func (apiCfg *apiConfig) handlerDeleteFeedFollow(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
) {
	feedFollowIDStr := chi.URLParam(r, "feedFollowID")

	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		respondWithError(
			w,
			http.StatusBadRequest,
			"Invalid feed follow ID",
		)
		return
	}

	deleteParams := database.DeleteFeedFollowParams{
		ID:     feedFollowID,
		UserID: user.ID,
	}

	err = apiCfg.DB.DeleteFeedFollow(
		r.Context(),
		deleteParams,
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't delete feed follow",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
