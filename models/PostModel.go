package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	Id 		primitive.ObjectID 	`bson:"_id"`
	UserId	primitive.ObjectID 	`bson:"user_id"`
	Text	string 				`bson:"text"`
	CopyId	primitive.ObjectID 	`bson:"copy_id"`
}