package books

import (
	"Sermire/APIServer/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	log "github.com/sirupsen/logrus"

	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"
)

func BooksRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("application/json"))

	r.Get("/", BookSearchHandler)
	r.Get("/{" + BOOKS_ID_KEY + "}", BookLookupHandler)
	r.Post("/", BookCreationHandler)

	return r
}

func GetBooksCollectionFromContext(c context.Context) *mongo.Collection {
	db, ok := c.Value("db").(*mongo.Client)
	if !ok {
		log.Panic("No database context found")
	}

	// Get the general sermire database
	SermireDB := db.Database(BOOKS_DB_NAME)
	// Get the books collection
	BooksCol := SermireDB.Collection(BOOKS_COL_NAME)

	return BooksCol
}

func BookLookupHandler(w http.ResponseWriter, r *http.Request) {
	// Get the book collection
	BookCol := GetBooksCollectionFromContext(r.Context())

	// Pull the book id from the URL
	bookIdCode := chi.URLParam(r, BOOKS_ID_KEY)
	bookId, err := primitive.ObjectIDFromHex(bookIdCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Pass the bookId through lookup to find the book
	book, err := BookLookup(BookCol, bookId)
	if err != nil {
		// Invalid book id
		log.WithFields(log.Fields{BOOKS_ID_KEY: bookId}).Info(err)
		w.WriteHeader(http.StatusNotFound)
	} else{
		var bookjson models.BookJSON

		bookjson.Id = book.Id.Hex()

		bookjson.Title = book.Title
		bookjson.Authors = book.Authors
		bookjson.PublicationYear = book.PublicationYear
		bookjson.RecordLanguage = book.RecordLanguage

		json.NewEncoder(w).Encode(bson.M{BOOK_KEY: bookjson})
	}
}

// TODO: Do database search (using $text) to get a list of possible books
func BookSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Get the book collection
	BookCol := GetBooksCollectionFromContext(r.Context())

	// Get the search string from request
	searchStr := r.FormValue(BOOK_SEARCH_STR_KEY)

	// Get the page number (default 1 if not found)
	var page int64 = 1
	var err error
	pageStr := r.FormValue(BOOK_SEARCH_PAGE_KEY)
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

	// Run the search
	books, err := BookSearch(BookCol, searchStr, MAX_SEARCH_RESULTS, page)
	if err != nil {
		json.NewEncoder(w).Encode(bson.M{BOOKS_KEY: nil})
		return
	}

	json.NewEncoder(w).Encode(bson.M{BOOKS_KEY: books})
}

// TODO: Create book in database
func BookCreationHandler(w http.ResponseWriter, r *http.Request) {
	// Get the book collection
	BookCol := GetBooksCollectionFromContext(r.Context())

	// Pull book info from request
	title := r.FormValue(BOOK_TITLE_KEY)
	// TODO: validate title

	author := r.FormValue(BOOK_AUTHOR_KEY)
	// TODO: validate author

	recordLang := r.FormValue(BOOK_RECORD_LANGUAGE_KEY)
	// TODO: validate record language

	publicationYearStr := r.FormValue(BOOK_PUBLICATION_YEAR_KEY)
	// TODO: validate publication year
	var publicationYear int64 = math.MinInt64
	var err error
	if publicationYearStr != "" {
		publicationYear, err = strconv.ParseInt(publicationYearStr, 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Soft publication date requirement to allow upcoming books to be put on the site before release
		// Lower bound is set around the time of the oldest book known (The Epic of Gilgamesh ~2100 BC)
		if publicationYear > (int64)(time.Now().Year()) + 1 || publicationYear < LOWEST_PUB_YEAR {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Attempt book insertion
	err = CreateBook(BookCol, title, author, publicationYear, recordLang)
	if err != nil {
		log.WithFields(log.Fields{BOOK_TITLE_KEY: title, BOOK_AUTHOR_KEY: author, BOOK_PUBLICATION_YEAR_KEY: publicationYear, BOOK_RECORD_LANGUAGE_KEY: recordLang}).Panic(err)
	}
}

func CreateBook(col *mongo.Collection, title string, author string, publicationYear int64, recordLang string) error {
	bookBSON := bson.D{{}}
	if publicationYear != math.MinInt64 {
		bookBSON = bson.D{{BOOKS_DB_TITLE_KEY, title}, {BOOKS_DB_AUTHOR_KEY, author}, {BOOKS_DB_PUBLICATION_YEAR_KEY, publicationYear}, {BOOKS_DB_RECORD_LANGUAGE_KEY, recordLang}}
	} else {
		bookBSON = bson.D{{BOOKS_DB_TITLE_KEY, title}, {BOOKS_DB_AUTHOR_KEY, author}, {BOOKS_DB_RECORD_LANGUAGE_KEY, recordLang}}
	}

	_, err := col.InsertOne(context.TODO(), bookBSON)
	return err
}

func BookSearch(col *mongo.Collection, searchText string, maxResults int64, page int64) ([]models.Book, error) {
	filter := bson.M{
		"$text": bson.D{{"$search", searchText}},
	}

	var booksResult []models.Book
	searchOptions := options.Find()
	searchOptions = searchOptions.SetLimit(maxResults)
	searchOptions = searchOptions.SetSkip((page - 1) * maxResults)
	ctx := context.TODO()
	cur, err := col.Find(ctx, filter, searchOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		if len(booksResult) >= (int)(maxResults) {
			break
		}
		var book models.Book

		err = cur.Decode(&book)
		if err != nil {
			return nil, err
		}

		booksResult = append(booksResult,book)
	}
	return booksResult, nil
}

func BookLookup(col *mongo.Collection, bookId primitive.ObjectID) (*models.Book, error) {
	var book models.Book

	singleRes := col.FindOne(context.TODO(), bson.M{BOOKS_DB_ID_KEY: bookId})
	err := singleRes.Decode(&book)

	return &book, err
}