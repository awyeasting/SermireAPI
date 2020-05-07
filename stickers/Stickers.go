package stickers

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	log "github.com/sirupsen/logrus"

	"Sermire/APIServer/models"
	"Sermire/APIServer/books"

	"context"
	"encoding/json"
	"net/http"
)
 
// Route all sticker related API handles to the proper handlers
func StickerRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("application/json"))

	r.Get("/{" + STICKER_CODE_KEY + "}", StickerLookupHandler)
	r.Post("/{" + STICKER_CODE_KEY + "}/{" + STICKER_BOOKID_KEY + "}", StickerUpdateHandler)

	return r
}

// Pulls the database client from context and then gets the sticker collection using the client
func GetStickerCollectionFromContext(c context.Context) *mongo.Collection {
	db, ok := c.Value("db").(*mongo.Client)
	if !ok {
		log.Panic("No database context found")
	}

	// Get the handle on the sticker collection based on information from the configuration
	SermireDB := db.Database(STICKERS_DB_NAME)
	StickerCol := SermireDB.Collection(STICKERS_COL_NAME)

	return StickerCol
}

// Performs sticker lookup on given code
func StickerLookupHandler(w http.ResponseWriter, r *http.Request) {
	// Get the stickers collection
	StickerCol := GetStickerCollectionFromContext(r.Context())

	// Pull the sticker code from the URL
	stickerCode := chi.URLParam(r, STICKER_CODE_KEY)

	// Pass the sticker code through lookup to find the sticker
	sticker, err := StickerLookup(StickerCol, stickerCode)
	if err != nil {
		// Invalid sticker code
		log.WithFields(log.Fields{STICKER_CODE_KEY: stickerCode}).Info(err)
		w.WriteHeader(http.StatusNotFound)
	} else{
		var stickerjson models.StickerJSON

		stickerjson.Id = sticker.Id.Hex()
		stickerjson.Code = sticker.Code
		if !sticker.BookId.IsZero() {
			// Sticker assigned
			stickerjson.BookId = sticker.BookId.Hex()
		} else {
			// Sticker not assigned
			stickerjson.BookId = nil
		}

		json.NewEncoder(w).Encode(bson.M{STICKER_KEY: stickerjson})
	}
}

// Update sticker in database to reflect a new status
func StickerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Get the stickers collection
	StickerCol := GetStickerCollectionFromContext(r.Context())

	// Pull the sticker code from the URL
	stickerCode := chi.URLParam(r, STICKER_CODE_KEY)

	// Check sticker code is valid and book is not set
	sticker, err := StickerLookup(StickerCol, stickerCode)
	if err != nil {
		// Invalid sticker code
		log.WithFields(log.Fields{STICKER_CODE_KEY: stickerCode}).Info(err)
		w.WriteHeader(http.StatusNotFound)
	} else if !sticker.BookId.IsZero() {
		// If the sticker book is set then it can only be changed by petition
		json.NewEncoder(w).Encode(bson.M{"error": STICKER_ALREADY_SET_ERROR_MSG})
	} else {
		// Check that book id is valid
		bookId := chi.URLParam(r, STICKER_BOOKID_KEY)
		bookPrimId, err := primitive.ObjectIDFromHex(bookId)
		if err != nil {
			json.NewEncoder(w).Encode(bson.M{"error": STICKERID_INVALID_ERROR_MSG})
			return
		}

		books.BookLookup(books.GetBooksCollectionFromContext(r.Context()), bookPrimId)
		// TODO

		// Set sticker book id
		err = StickerUpdate(StickerCol, sticker, bookPrimId)
		if err != nil {
			log.WithFields(log.Fields{STICKER_KEY: sticker, "bookPrimId": bookPrimId.Hex()}).Panic(err)
		}

		json.NewEncoder(w).Encode(bson.M{"result": "success"})
	}
}

// Looks up the sticker code in the database
func StickerLookup(col *mongo.Collection, stickerCode string) (*models.Sticker, error) {
	var sticker models.Sticker

	singleRes := col.FindOne(context.TODO(), bson.M{STICKER_DB_CODE_KEY: stickerCode})
	err := singleRes.Decode(&sticker)

	return &sticker, err
}

func StickerIDLookup(col *mongo.Collection, stickerID primitive.ObjectID) (*models.Sticker, error) {
	var sticker models.Sticker

	singleRes := col.FindOne(context.TODO(), bson.M{"_id": stickerID})
	err := singleRes.Decode(&sticker)

	return &sticker, err
}

// Updates a sticker to point to the given book id
func StickerUpdate(col *mongo.Collection, sticker *models.Sticker, bookId primitive.ObjectID) error {
	filter := bson.M{"_id": sticker.Id}
	update := bson.D{
		{"$set", bson.D{
			{"book_id",bookId},
		}},
	}
	_, err := col.UpdateOne(context.TODO(), filter, update)
	return err
}