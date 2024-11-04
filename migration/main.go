package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Manga struct {
	ID     string   `json:"id" bson:"_id,omitempty"`
	Title  string   `json:"title" bson:"title"`
	Genres []string `json:"genres" bson:"genres"`
}

type User struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Email string `json:"email" bson:"email"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	neo4jURI := os.Getenv("NEO4J_URI")
	neo4jUsername := os.Getenv("NEO4J_USERNAME")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	fmt.Println(mongoURI)
	fmt.Println(neo4jURI)
	fmt.Println(neo4jUsername)
	fmt.Println(neo4jPassword)

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())
	mangaCollection := mongoClient.Database("manga_store").Collection("manga")
	userCollection := mongoClient.Database("manga_store").Collection("users")

	// Connect to Neo4j
	driver, err := neo4j.NewDriver(neo4jURI, neo4j.BasicAuth(neo4jUsername, neo4jPassword, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()

	// Insert Manga nodes
	cursor, err := mangaCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var manga Manga
		if err := cursor.Decode(&manga); err != nil {
			log.Fatal(err)
		}

		session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err = session.Run(
			"MERGE (m:Manga {id: $id, title: $title, genres: $genres})",
			map[string]interface{}{
				"id":     manga.ID,
				"title":  manga.Title,
				"genres": manga.Genres,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		session.Close()
	}

	// Insert User nodes
	cursor, err = userCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}

		session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err = session.Run(
			"MERGE (u:User {id: $id, email: $email})",
			map[string]interface{}{
				"id":    user.ID,
				"email": user.Email,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		session.Close()
	}

	fmt.Println("Data migration to Neo4j completed successfully!")
}
