package models 

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	UserId		primitive.ObjectID 	`bson:"_id" json:"id"`
	Email		string 				`bson:"email" json:"email"`
	Username 	string 				`bson:"username" json:"username"`
	FirstName 	string 				`bson:"firstname" json:"firstname"`
	LastName 	string 				`bson:"lastname" json:"lastname"`
	Password 	string 				`bson:"password" json:"password"`
	Token 		string 				`bson:"token" json:"token"`
}

type InsertUser struct {
	Email		string 				`bson:"email" json:"email"`
	Username 	string 				`bson:"username" json:"username"`
	FirstName 	string 				`bson:"firstname" json:"firstname"`
	LastName 	string 				`bson:"lastname" json:"lastname"`
	Password 	string 				`bson:"password" json:"password"`
	Token 		string 				`bson:"token" json:"token"`
}

type PublicUser struct {
	Username 	string `json:"username"`
	FirstName 	string `json:"firstname"`
	LastName 	string `json:"lastname"`
}

type ResponseResult struct {
	Error		string `json:"error"`
	Result 		string `json:"result"`
}