package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sticker struct {
	Id 		primitive.ObjectID 	`bson:"_id"`
	Code	string 				`bson:"code"`
	BookId	primitive.ObjectID 	`bson:"book_id"`
}

type StickerJSON struct {
	Id 		string 		`json:"_id"`
	Code	string 		`json:"code"`
	BookId	interface{} `json:"book_id"`
}