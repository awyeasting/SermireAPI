package login 

import (
	"Sermire/APIServer/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	log "github.com/sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"

	"context"
	"encoding/json"
	"net/http"
)

func LoginRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("application/json"))

	r.With(DecodeUserInfo).Post("/", LoginHandler)
	r.With(DecodeUserInfo).With(ValidatePassword).Post("/register", RegisterHandler)
	r.Get("/password-policy",ServePasswordPolicy)
	
	return r
}

func GetLoginCollectionFromContext(c context.Context) *mongo.Collection {
	db, ok := c.Value("db").(*mongo.Client)
	if !ok {
		log.Panic("No database context found")
	}

	// Get the general sermire database
	SermireDB := db.Database(LOGIN_DB_NAME)
	// Get the books collection
	BooksCol := SermireDB.Collection(USER_COL_NAME)

	return BooksCol
}

func ServePasswordPolicy(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w,r,"password_policy.txt")
}

// Handler function for a login attempt
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// Get user information
	user := r.Context().Value("user")
	collection := GetLoginCollectionFromContext(r.Context())

	var result models.User
	// Find the first (and only) user's username in the collection
	err := collection.FindOne(context.TODO(), bson.D{{"username", user.(models.User).Username}}).Decode(&result)

	if err != nil {
		// Hash to keep response time in line with a check on an existing user
		bcrypt.GenerateFromPassword([]byte(user.(models.User).Password), bcrypt.DefaultCost)
		WriteJSONResponse(w, bson.M{"result": "fail", "error": "Username or password not found"}, http.StatusBadRequest)
		return
	}

	// Check that the password hash matches the entered password
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.(models.User).Password))
	// If the hash doesn't match then fail it
	if err != nil {
		WriteJSONResponse(w, bson.M{"result": "fail", "error": "Username or password not found"}, http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  result.Username,
		"firstname": result.FirstName,
		"lastname":  result.LastName,
	})

	// TODO: Change secret
	tokenString, err := token.SignedString([]byte("secret"))

	var res models.ResponseResult
	if err != nil {
		res.Error = "Error while generating token, Try again"
		json.NewEncoder(w).Encode(res)
		return
	}

	result.Token = tokenString
	result.Password = ""

	json.NewEncoder(w).Encode(result)
}

// Handler function for a user registration attempt
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Received register request")

	// Get user information that was passed in via json
	user := (r.Context().Value("user")).(models.User)

	collection := GetLoginCollectionFromContext(r.Context())

	// Search for username in database
	var result models.User
	err := collection.FindOne(context.TODO(), bson.D{{"username", user.Username}}).Decode(&result)
	if err != nil {
		// Check if the username is not taken
		if err.Error() == "mongo: no documents in result" {
			// Hash the entered password
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				WriteJSONResponse(w, bson.M{"result":"fail", "error": "Error while hashing password, try again"}, http.StatusInternalServerError)
				log.WithFields(log.Fields{"error": err.Error()}).Error("Failed to hash password for unknown reason")
				return
			}
			user.Password = string(hash)

			// Try to create the user in the database
			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				WriteJSONResponse(w, bson.M{"result": "fail", "error": "Error creating user account, try again"}, http.StatusInternalServerError)
				log.WithFields(log.Fields{"error": err.Error()}).Error("Failed to register user account for unknown reason")
				return
			}
			// If it didn't fail then it succeded 
			WriteJSONResponse(w, bson.M{"result": "success"}, http.StatusOK)
			log.Info("New user registered")
			return
		}

		// Otherwise something is wrong
		WriteJSONResponse(w, bson.M{"result": "fail", "error": "Error looking up username, try again"}, http.StatusInternalServerError)
		log.WithFields(log.Fields{"error": err.Error()}).Error("Failed to look up username, but not because no documents.")
		return
	}

	// Case: Username taken
	WriteJSONResponse(w, bson.M{"result": "fail", "error": "User already exists"}, http.StatusConflict)
	log.Info("Registration conflict")
}

// Writes a given interface to an http ResponseWriter with a given status code
func WriteJSONResponse(w http.ResponseWriter, jsonResponse interface{}, header int) {
	json.NewEncoder(w).Encode(jsonResponse)
}
