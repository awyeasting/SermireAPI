package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	Id 		primitive.ObjectID 	`bson:"_id"`
	User	string 				`bson:"user"`
	Text	string 				`bson:"text"`
	Tags	[]string			`bson:"tags"`
	Date	primitive.DateTime	`bson:"date"`
	CopyId	primitive.ObjectID 	`bson:"copy"`
}