package posts

import (
	"github.com/go-chi/chi"
	//"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	log "github.com/sirupsen/logrus"

	"SermireAPI/stickers"
	"SermireAPI/models"
	"SermireAPI/login"

	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func PostsRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/copy/{" + POSTS_COPY_KEY + "}", GetCopyPostsHandler)
	r.Get("/book/{" + POSTS_BOOK_KEY + "}", GetBookPostsHandler)

	r.Post("/{" + POSTS_COPY_KEY + "}", CopyPostHandler)

	return r
}

// Pulls the database client from context and then gets the posts collection using the client
func getPostsCollectionFromContext(c context.Context) *mongo.Collection {
	db, ok := c.Value("db").(*mongo.Client)
	if !ok {
		log.Panic("No database context found")
	}

	// Get the handle on the posts collection based on information from the configuration
	SermireDB := db.Database(POSTS_DB_NAME)
	PostsCol := SermireDB.Collection(POSTS_COL_NAME)

	return PostsCol
}

func GetCopyPostsHandler(w http.ResponseWriter, r *http.Request) {
	PostsCol := getPostsCollectionFromContext(r.Context())

	stickerCode := chi.URLParam(r, POSTS_COPY_KEY)
	
	sticker, err := stickers.StickerLookup(stickers.GetStickerCollectionFromContext(r.Context()), stickerCode)
	if err != nil {
		log.WithFields(log.Fields{"stickerCode": stickerCode}).Panic(err)
	}

	// Get the page number (default 1 if not found)
	// TODO provide information on how many pages there are
	var page int64 = 1
	pageStr := r.FormValue(POSTS_PAGE_KEY)
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if page < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	posts, err := GetCopyPosts(PostsCol, sticker.Id, MAX_POSTS_RESULTS, page)
	if err != nil {
		log.WithFields(log.Fields{"stickerCode": stickerCode}).Panic(err)
	}

	json.NewEncoder(w).Encode(bson.M{POSTS_KEY: posts})
}

func GetBookPostsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func CopyPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Received post request")
	PostsCol := getPostsCollectionFromContext(r.Context())

	// Get user mongo ID from token
	userToken := r.FormValue(POSTS_TOKEN_KEY)
	claims, err := login.GetClaims(userToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := claims["username"].(string)

	text := r.FormValue(POSTS_TEXT_KEY)
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stickerCode := chi.URLParam(r, POSTS_COPY_KEY)
	sticker, err := stickers.StickerLookup(stickers.GetStickerCollectionFromContext(r.Context()), stickerCode)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Info(err)
		return
	}

	tags := strings.Split(r.FormValue(POSTS_TAGS_KEY),";")

	err = CreatePost(PostsCol, username, text, tags, sticker.Id)
	if err != nil {
		log.WithFields(log.Fields{"stickerCode": stickerCode}).Panic(err)
	}
}

func CreatePost(col *mongo.Collection, user string, text string, tags []string, copyId primitive.ObjectID) error {
	postBSON := bson.D{{POSTS_DB_COPY_KEY, copyId}, 
		{POSTS_DB_TEXT_KEY, text}, 
		{POSTS_DB_TAGS_KEY, tags},
		{POSTS_DB_DATE_KEY, primitive.NewDateTimeFromTime(time.Now())}, 
		{POSTS_DB_USER_KEY, user}}

	_, err := col.InsertOne(context.TODO(), postBSON)
	return err
}

func GetCopyPosts(col *mongo.Collection, copyId primitive.ObjectID, maxResults int64, page int64) ([]models.Post, error) {
	filter := bson.M{POSTS_DB_COPY_KEY: copyId}

	var postsResult []models.Post
	searchOptions := options.Find()
	searchOptions = searchOptions.SetLimit(maxResults)
	// Not optimal for large skip numbers
	searchOptions = searchOptions.SetSkip((page - 1) * maxResults)
	searchOptions = searchOptions.SetSort(bson.D{{POSTS_DB_DATE_KEY,-1}})
	ctx := context.TODO()
	cur, err := col.Find(ctx, filter, searchOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		if len(postsResult) >= (int)(maxResults) {
			break
		}
		var post models.Post

		err = cur.Decode(&post)
		if err != nil {
			return nil, err
		}

		postsResult = append(postsResult, post)
	}
	return postsResult, nil
}