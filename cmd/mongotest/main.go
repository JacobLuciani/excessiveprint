package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dburl, ok := os.LookupEnv("DB_URL")
	if !ok {
		log.Fatal("DB_URL must be set in environment or in .env file")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dburl))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	coll := client.Database("testdb").Collection("names")
	name := map[string]interface{}{
		"value": "Jacob",
	}

	result, err := coll.InsertOne(context.TODO(), name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted at ", result.InsertedID)
}
