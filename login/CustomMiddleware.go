package login

import (
	"SermireAPI/models"

	"go.mongodb.org/mongo-driver/bson"
	log "github.com/sirupsen/logrus"

	"context"
	"encoding/json"
	"io/ioutil"	
	"net/http"
)

// Custom middleware for taking in user info
func DecodeUserInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unmarshal json from request body
		var user models.User
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &user)
		var res models.ResponseResult
		if err != nil {
			WriteJSONResponse(w, res, http.StatusBadRequest)
			log.Error(err)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Custom middleware for checking if a password matches required criteria
func ValidatePassword(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(models.User)

		// Pull user password and check that is matches length criteria
		pass := user.Password
		if len(pass) < 8 {
			WriteJSONResponse(w, bson.M{"result": "fail", "error": "Password is too short"}, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}