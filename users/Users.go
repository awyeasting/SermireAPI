package users

import (
	"SermireAPI/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	log "github.com/sirupsen/logrus"

	"context"
	"encoding/json"
	"net/http"
)

func UsersRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("application/json"))

	r.Get("/{" + USERS_ID_KEY + "}", UserLookupHandler)

	return r
}

func GetUsersCollectionFromContext(c context.Context) *mongo.Collection {
	db, ok := c.Value("db").(*mongo.Client)
	if !ok {
		log.Panic("No database context found")
	}

	// Get the general sermire database
	SermireDB := db.Database(USERS_DB_NAME)
	// Get the books collection
	UsersCol := SermireDB.Collection(USERS_COL_NAME)

	return UsersCol
}

func UserLookupHandler(w http.ResponseWriter, r *http.Request) {
	// Get the users Collection
	UsersCol := GetUsersCollectionFromContext(r.Context())

	// Pull the user id from the URL
	userIdCode := chi.URLParam(r, USERS_ID_KEY)
	userId, err := primitive.ObjectIDFromHex(userIdCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Pass the user id through lookup to find the user
	user, err := UserLookup(UsersCol, userId)
	if err != nil {
		// Invalid user id
		log.WithFields(log.Fields{USERS_ID_KEY: userId}).Info(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		json.NewEncoder(w).Encode(bson.M{USER_KEY: user})
	}
}

// Looks up the user id in the database
func UserLookup(col *mongo.Collection, userId primitive.ObjectID) (*models.PublicUser, error) {
	var user models.PublicUser

	singleRes := col.FindOne(context.TODO(), bson.M{USERS_DB_ID_KEY: userId})
	err := singleRes.Decode(&user)

	return &user, err
}