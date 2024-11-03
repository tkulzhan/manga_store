package databases

import (
	"context"
	"manga_store/internal/logger"
	"manga_store/internal/helpers"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func InitMongo() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file: " + err.Error())
	}

	mongoUri := helpers.GetEnv("MONGO_URI", "mongodb://127.0.0.1:27017")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoUri)

	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error(err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Connected to MongoDB")
}

func Users() *mongo.Collection {
	return client.Database("manga_store").Collection("users")
}

func Manga() *mongo.Collection {
	return client.Database("manga_store").Collection("users")
}

func Activities() *mongo.Collection {
	return client.Database("manga_store").Collection("activities")
}