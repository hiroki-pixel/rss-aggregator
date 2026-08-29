package main

import (
	"net/http"
	"rss-aggregator/internal/database"
)

// authedHandler = 認証済みUserも受け取れるHandlerの型.関数の形そのものに名前を付けている
type authedHandler func(
	http.ResponseWriter,
	*http.Request,
	database.User,
)

func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := getAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Couldn't find API key")
			return
		}

		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't get user")
			return
		}

		handler(w, r, user)
	}
}
