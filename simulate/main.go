package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Email string `json:"email" bson:"email"`
}

type Manga struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	fmt.Println("Mongo URI:", mongoURI)

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())
	userCollection := mongoClient.Database("manga_store").Collection("users")
	mangaCollection := mongoClient.Database("manga_store").Collection("manga")

	// Fetch all users
	cursor, err := userCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	// Fetch all manga
	var mangas []Manga
	mangaCursor, err := mangaCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer mangaCursor.Close(context.Background())

	for mangaCursor.Next(context.Background()) {
		var manga Manga
		if err := mangaCursor.Decode(&manga); err != nil {
			log.Fatal(err)
		}
		mangas = append(mangas, manga)
	}

	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}

		// Step 1: Login
		cookies, err := loginUser(user.Email, "12345678")
		if err != nil {
			log.Printf("Error logging in user %s: %v", user.Email, err)
			continue
		}

		// Select a random manga
		rand.Seed(time.Now().UnixNano())
		randomManga := mangas[rand.Intn(len(mangas))]

		// Step 2: Get Manga details
		err = getMangaDetails(cookies, randomManga.ID)
		if err != nil {
			log.Printf("Error fetching manga details for user %s: %v", user.Email, err)
			continue
		}

		// Step 3: Purchase Manga
		err = purchaseManga(cookies, randomManga.ID)
		if err != nil {
			log.Printf("Error purchasing manga for user %s: %v", user.Email, err)
			continue
		}

		// Step 4: Rate Manga
		randomRating := float64(rand.Intn(3)+3) + rand.Float64()
		err = rateManga(cookies, randomManga.ID, randomRating)
		if err != nil {
			log.Printf("Error rating manga for user %s: %v", user.Email, err)
			continue
		}

		log.Printf("User %s successfully completed all actions", user.Email)
	}
}

func loginUser(email, password string) ([]*http.Cookie, error) {
	url := "http://localhost:3000/auth/login"
	body := map[string]string{"email": email, "password": password}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	return resp.Cookies(), nil
}

func getMangaDetails(cookies []*http.Cookie, mangaID string) error {
	url := fmt.Sprintf("http://localhost:3000/manga/%s", mangaID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get manga details failed with status code: %d", resp.StatusCode)
	}

	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func purchaseManga(cookies []*http.Cookie, mangaID string) error {
	url := "http://localhost:3000/manga/purchase"
	body := map[string]string{"mangaId": mangaID}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("purchase manga failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func rateManga(cookies []*http.Cookie, mangaID string, score float64) error {
	url := fmt.Sprintf("http://localhost:3000/manga/%s/rate", mangaID)
	body := map[string]float64{"score": score}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rate manga failed with status code: %d", resp.StatusCode)
	}

	return nil
}
