package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"rss-aggregator/internal/database"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	// PostgreSQLドライバをdatabase/sqlに登録するためのblank import
	_ "github.com/lib/pq"
)

// handlerReadiness はアプリケーションがHTTPリクエストを
// 受け付けられる状態であることを確認するためのヘルスチェックHandler。
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, struct{}{})
}

// handlerHello は動作確認用のHandler。
func handlerHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

// handlerErr はエラーレスポンスの動作確認用Handler。
func handlerErr(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusBadRequest, "Something went wrong")
}

// handlerCreateUser はリクエストBodyからユーザー名を受け取り、
// 新しいユーザーをDBに作成する。
func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}

	// リクエストBodyのJSONをGoの構造体へ変換する。
	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Printf("Failed to decode request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	// DBへINSERTするためのパラメータを組み立てる。
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
	}

	// sqlcが生成したCreateUserを使ってユーザーをDBへ登録する。
	user, err := apiCfg.DB.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	respondWithJson(w, http.StatusCreated, databaseUserToUser(user))
}

// handlerGetUser は認証Middlewareから渡されたユーザー情報を返す。
// APIキーの検証とユーザー取得はmiddlewareAuth側で完了している。
func (apiCfg *apiConfig) handlerGetUser(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
) {
	respondWithJson(w, http.StatusOK, databaseUserToUser(user))
}

// handlerGetFeeds は認証不要
func (apiCfg apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := apiCfg.DB.GetFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get feeds")
		return
	}

	respondWithJson(w, http.StatusOK, databaseFeedsToFeeds(feeds))
}

// databaseUserToUser はDB用のUserモデルをAPIレスポンス用のUserへ変換する。
// DBの内部表現と外部公開するJSONを分離するために変換処理を挟んでいる。
func databaseUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		ApiKey:    dbUser.ApiKey,
	}
}

// databaseFeedToFeed はDB用のFeedモデルをAPIレスポンス用のFeedへ変換する。
func databaseFeedToFeed(dbFeed database.Feed) Feed {
	return Feed{
		ID:        dbFeed.ID,
		CreatedAt: dbFeed.CreatedAt,
		UpdatedAt: dbFeed.UpdatedAt,
		Name:      dbFeed.Name,
		Url:       dbFeed.Url,
		UserID:    dbFeed.UserID,
	}
}

// databaseFeedsToFeedsはDB用の対象ユーザのFeeds全件返却モデルをレスポンス用のFeedsへ変換する
func databaseFeedsToFeeds(dbFeeds []database.Feed) []Feed {
	feeds := make([]Feed, 0, len(dbFeeds))

	for _, dbFeed := range dbFeeds {
		feeds = append(feeds, databaseFeedToFeed(dbFeed))
	}
	return feeds

}

func databaseFeedFollowToFeedFollow(dbFeedFollow database.FeedFollow) FeedFollow {
	return FeedFollow{
		ID:        dbFeedFollow.ID,
		CreatedAt: dbFeedFollow.CreatedAt,
		UpdatedAt: dbFeedFollow.UpdatedAt,
		UserID:    dbFeedFollow.UserID,
		FeedID:    dbFeedFollow.FeedID,
	}
}

func databaseFeedFollowsToFeedFollws(dbFeedFollows []database.FeedFollow) []FeedFollow {
	feedFollows := make([]FeedFollow, 0, len(dbFeedFollows))

	for _, dbFeedFollow := range dbFeedFollows {
		feedFollows = append(feedFollows, databaseFeedFollowToFeedFollow(dbFeedFollow))
	}
	return feedFollows
}

func databasePostToPost(dbPost database.Post) Post {
	return Post{
		ID:          dbPost.ID,
		CreatedAt:   dbPost.CreatedAt,
		UpdatedAt:   dbPost.UpdatedAt,
		Title:       dbPost.Title,
		Url:         dbPost.Url,
		Description: dbPost.Description,
		PublishedAt: dbPost.PublishedAt,
		FeedID:      dbPost.FeedID,
	}
}

func databasePostsToPosts(dbPosts []database.Post) []Post {
	posts := make([]Post, 0, len(dbPosts))

	for _, dbPost := range dbPosts {
		posts = append(posts, databasePostToPost(dbPost))
	}
	return posts
}

// apiConfig はHandler間で共有するアプリケーション依存関係を保持する。
// 現在はsqlcが生成したDB操作用Queriesを保持している。
type apiConfig struct {
	DB *database.Queries
}

func main() {
	// .envからアプリケーション設定を読み込む。
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// HTTPサーバがListenするポートを取得する。
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not found")
	}

	// PostgreSQLへの接続情報を取得する。
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found")
	}

	// database/sqlを通してPostgreSQLを操作するためのDBハンドルを作成する。
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// アプリ起動時にDBへ実際に接続できることを確認する。
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database")

	// *sql.DBをsqlcが生成したQueriesで包み、
	// CreateUserやCreateFeedなどのアプリ固有のDB操作を利用可能にする。
	dbQueries := database.New(db)

	// HandlerからDB操作を利用できるよう依存関係をまとめる。
	apiCfg := apiConfig{
		DB: dbQueries,
	}

	// HTTPリクエストのURL・Methodを各Handlerへ振り分けるRouterを作成する。
	router := chi.NewRouter()
	v1Router := chi.NewRouter()

	// /v1配下のAPIルート。
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)

	v1Router.Post("/users", apiCfg.handlerCreateUser)

	// 認証が必要なAPIにはmiddlewareAuthを適用する。
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))

	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))

	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	// v1Routerを/v1配下へまとめて登録する。
	router.Mount("/v1", v1Router)

	// APIバージョン外の動作確認用ルート。
	router.Get("/hello", handlerHello)

	//  動作確認
	go startScraping(&apiCfg, 10*time.Second, 3)

	v1Router.Get(
		"/posts",
		apiCfg.middlewareAuth(apiCfg.handlerGetPosts),
	)

	// RouterをHTTPサーバへ設定する。
	server := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}

	log.Printf("Server listening on port %s", port)

	// HTTPサーバを起動し、リクエストの待受を開始する。
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
