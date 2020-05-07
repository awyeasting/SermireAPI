package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	Id 				primitive.ObjectID 	`bson:"_id"`
	Title			string				`bson:"title"`
	Authors			string				`bson:"author"`
	PublicationYear	int 				`bson:"publication_year"`
	RecordLanguage	string				`bson:"record_language"`
	Original		interface{}			`bson:"original"`
}

type BookJSON struct {
	Id 				string 		`json:"_id"`
	Title			string		`json:"title"`
	Authors			string		`json:"author"`
	PublicationYear	int			`json:"publication_year"`
	RecordLanguage	string		`json:"record_language"`
	Original		interface{}	`json:"original"`
}