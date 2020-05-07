package db

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"
	"fmt"
	"time"
)

func GetDBClient(conn_info string) (*mongo.Client, error) {

	// Create a new client for the database
	client, err := mongo.NewClient(options.Client().ApplyURI(conn_info))
	if err != nil {
		return nil, err
	}

	// Attempt to connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Client failed to connect")
		return nil, err
	}

	// Check that connection is usable
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("Ping failed.")
		return nil, err
	}

	// Success
	return client, nil
}