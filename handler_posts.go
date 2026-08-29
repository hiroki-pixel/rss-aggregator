package main

import (
	"net/http"
	"rss-aggregator/internal/database"
)

func (apiCfg *apiConfig) handlerGetPosts(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
) {
	posts, err := apiCfg.DB.GetPostsForUser(
		r.Context(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  10,
		},
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't get posts",
		)
		return
	}

	respondWithJson(
		w,
		http.StatusOK,
		databasePostsToPosts(posts),
	)
}
