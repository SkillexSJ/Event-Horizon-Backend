package db

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectDB() *mongo.Database {
	MongoURI := os.Getenv("MONGO_URI")

	if MongoURI == "" {
		// EXIT IF MONGO_URI IS NOT SET
		log.Fatal("MONGO_URI environment variable not set")
	}

	clientOptions := options.Client().ApplyURI(MongoURI)

	client, err := mongo.Connect(clientOptions)

	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}

	log.Println("Connected to MongoDB")

	return client.Database(os.Getenv("DATABASE_NAME"))

}
