package services

import (
	"context"
	"encoding/json"
	"errors"
	"manga_store/internal/databases"
	"manga_store/internal/logger"
	"manga_store/internal/models"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MangaService struct {
	manga *mongo.Collection
	users *mongo.Collection
	redis *redis.Client
	neo4j neo4j.SessionWithContext
}

var mu = sync.Mutex{}

func NewMangaService() MangaService {
	s := MangaService{
		manga: databases.Manga(),
		users: databases.Users(),
		redis: databases.Redis(),
		neo4j: databases.Neo4j(context.Background()),
	}

	go func(s MangaService) {
		for {
			if err := s.updatePopularMangaCache(); err != nil {
				logger.Error("Erro updating popular manga cache")
			}
			time.Sleep(time.Minute)
		}
	}(s)

	return s
}

func (s MangaService) GetNewestManga(limit int) ([]models.Manga, error) {
	ctx := context.Background()
	var mangas []models.Manga

	findOptions := options.Find()
	findOptions.SetSort(map[string]interface{}{"createdAt": -1})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := s.manga.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var manga models.Manga
		if err := cursor.Decode(&manga); err != nil {
			return nil, err
		}
		mangas = append(mangas, manga)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return mangas, nil
}

func (s MangaService) CreateManga(title, author, description string, price float64, quantity int, genres []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	manga := models.Manga{
		Title:       title,
		Author:      author,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		Genres:      genres,
		CreatedAt:   int(time.Now().Unix()),
	}

	result, err := s.manga.InsertOne(ctx, manga)
	if err != nil {
		return err
	}

	mangaID := result.InsertedID.(primitive.ObjectID).Hex()

	neo4jCtx := context.Background()
	_, err = s.neo4j.ExecuteWrite(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(neo4jCtx, `
			CREATE (m:Manga {id: $id, title: $title, genres: $genres})
		`, map[string]interface{}{
			"id":     mangaID,
			"title":  title,
			"genres": genres,
		})
		return nil, err
	})

	if err != nil {

		s.manga.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
		return errors.New("failed to create manga in Neo4j, creation rolled back")
	}

	return nil
}

func (s MangaService) DeleteManga(mangaID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{"$set": bson.M{"isDeleted": true}})
	if err != nil {
		return err
	}

	neo4jCtx := context.Background()
	_, err = s.neo4j.ExecuteWrite(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(neo4jCtx, `
			MATCH (m:Manga {id: $id})
			DETACH DELETE m
		`, map[string]interface{}{
			"id": mangaID.Hex(),
		})
		return nil, err
	})

	if err != nil {
		return errors.New("failed to delete manga in Neo4j")
	}

	return nil
}

func (s MangaService) SearchManga(query string, genres []string, author string, limit int) ([]models.Manga, error) {
	ctx := context.Background()
	var mangas []models.Manga

	filter := bson.M{}
	if query != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		}
	}
	if len(genres) > 0 {
		filter["genres"] = bson.M{"$all": genres}
	}
	if author != "" {
		filter["author"] = bson.M{"$regex": author, "$options": "i"}
	}

	findOptions := options.Find().SetSort(bson.M{"createdAt": -1})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := s.manga.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var manga models.Manga
		if err := cursor.Decode(&manga); err != nil {
			return nil, err
		}
		mangas = append(mangas, manga)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return mangas, nil
}

func (s MangaService) GetMangaByID(id, userID string) (*models.Manga, error) {
	ctx := context.Background()
	var manga models.Manga

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectId}

	err = s.manga.FindOne(ctx, filter).Decode(&manga)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()

	update := bson.M{"$inc": bson.M{"views": 1}}
	_, err = s.manga.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	if err := s.createOrUpdateViewInNeo4j(userID, id, manga.Title, manga.Genres); err != nil {
		return nil, err
	}
	

	return &manga, nil
}

func (s MangaService) createOrUpdateViewInNeo4j(userID, mangaID string, title string, genres []string) error {
	ctx := context.Background()

	_, err := s.neo4j.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, `
			MERGE (u:User {id: $userID})
			ON CREATE SET u.id = $userID
			MERGE (m:Manga {id: $mangaID})
			ON CREATE SET m.title = $title, m.genres = $genres
			MERGE (u)-[:VIEWED]->(m)
		`, map[string]interface{}{
			"userID":  userID,
			"mangaID": mangaID,
			"title":   title,
			"genres":  genres,
		})
		return nil, err
	})

	return err
}

func (s MangaService) PurchaseManga(userID, mangaID primitive.ObjectID) error {
	ctx := context.Background()

	var manga models.Manga
	err := s.manga.FindOne(ctx, bson.M{"_id": mangaID, "isDeleted": false}).Decode(&manga)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("manga not found")
		}
		return err
	}

	if manga.Quantity <= 0 {
		return errors.New("manga is out of stock")
	}

	var user models.User
	err = s.users.FindOne(ctx, bson.M{"_id": userID, "isDeleted": false}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user not found")
		}
		return err
	}

	purchase := models.Purchase{
		MangaID:      manga.ID,
		Title:        manga.Title,
		Price:        manga.Price,
		PurchaseDate: time.Now().Format(time.RFC3339),
	}
	userUpdate := bson.M{"$push": bson.M{"purchaseHistory": purchase}}

	_, err = s.users.UpdateOne(ctx, bson.M{"_id": userID}, userUpdate)
	if err != nil {
		mu.Unlock()
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	mangaUpdate := bson.M{"$inc": bson.M{"quantity": -1, "sold": 1}}
	_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, mangaUpdate)
	if err != nil {
		return err
	}

	if err := s.createOrUpdatePurchaseInNeo4j(userID, mangaID, manga.Title, manga.Genres); err != nil {
		return err
	}
	
	return err
}

func (s MangaService) createOrUpdatePurchaseInNeo4j(userID, mangaID primitive.ObjectID, title string, genres []string) error {
	ctx := context.Background()

	_, err := s.neo4j.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, `
			MERGE (u:User {id: $userID})
			ON CREATE SET u.id = $userID
			MERGE (m:Manga {id: $mangaID})
			ON CREATE SET m.title = $title, m.genres = $genres
			MERGE (u)-[:PURCHASED]->(m)
		`, map[string]interface{}{
			"userID":  userID.Hex(),
			"mangaID": mangaID.Hex(),
			"title":   title,
			"genres":  genres,
		})
		return nil, err
	})

	return err
}


func (s MangaService) GetPopularManga() ([]models.Manga, error) {
	ctx := context.Background()

	var mangas []models.Manga

	data, err := s.redis.Get(ctx, "popular_manga").Result()
	if err != nil {
		logger.Error("Error getting popular manga from cache, retrieving from db")
		mangas, err := s.getPopularMangaFromMongo()
		if err != nil {
			return mangas, nil
		}
	} else if err := json.Unmarshal([]byte(data), &mangas); err == nil {
		return mangas, nil
	}

	return nil, errors.New("failed to retrieve popular manga from cache")
}

func (s MangaService) RateManga(userID, mangaID primitive.ObjectID, rating float64) error {
	ctx := context.Background()

	var user models.User
	err := s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	var manga models.Manga
	err = s.manga.FindOne(ctx, bson.M{"_id": mangaID}).Decode(&manga)
	if err != nil {
		return err
	}

	for _, r := range user.Ratings {
		if r.MangaID == mangaID.Hex() {
			if err := s.updateExistingRating(userID, mangaID, rating); err != nil {
				return err
			}
			return s.createOrUpdateRatingInNeo4j(userID, mangaID, rating, manga.Title, manga.Genres)
		}
	}

	_, err = s.users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$push": bson.M{"ratings": models.Rating{
		MangaID: mangaID.Hex(),
		Score:   rating,
	}}})
	if err != nil {
		return err
	}

	if err := s.updateMangaRating(mangaID, rating, 1); err != nil {
		return err
	}

	return s.createOrUpdateRatingInNeo4j(userID, mangaID, rating, manga.Title, manga.Genres)
}

func (s MangaService) createOrUpdateRatingInNeo4j(userID, mangaID primitive.ObjectID, rating float64, title string, genres []string) error {
	ctx := context.Background()

	_, err := s.neo4j.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, `
			MERGE (u:User {id: $userID})
			ON CREATE SET u.id = $userID
			MERGE (m:Manga {id: $mangaID})
			ON CREATE SET m.title = $title, m.genres = $genres
			MERGE (u)-[r:RATED]->(m)
			SET r.score = $score
		`, map[string]interface{}{
			"userID":  userID.Hex(),
			"mangaID": mangaID.Hex(),
			"title":   title,
			"genres":  genres,
			"score":   rating,
		})
		return nil, err
	})

	return err
}

func (s MangaService) updateExistingRating(userID, mangaID primitive.ObjectID, newRating float64) error {
	ctx := context.Background()

	_, err := s.users.UpdateOne(ctx, bson.M{"_id": userID, "ratings.mangaId": mangaID},
		bson.M{"$set": bson.M{"ratings.$.score": newRating}})
	if err != nil {
		return err
	}

	var manga models.Manga
	err = s.manga.FindOne(ctx, bson.M{"_id": mangaID}).Decode(&manga)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	return s.updateMangaRating(mangaID, newRating, 0)
}

func (s MangaService) updateMangaRating(mangaID primitive.ObjectID, newRating float64, additionalRatedTimes int) error {
	ctx := context.Background()

	var manga models.Manga
	err := s.manga.FindOne(ctx, bson.M{"_id": mangaID}).Decode(&manga)
	if err != nil {
		return err
	}

	newRatedTimes := manga.RatedTimes + additionalRatedTimes
	if newRatedTimes == 1 {
		newTotalRating := newRating

		_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$set": bson.M{"rating": newTotalRating, "ratedTimes": newRatedTimes},
		})
		if err != nil {
			return err
		}
	} else if newRatedTimes > 0 {
		newTotalRating := ((manga.Rating * float64(manga.RatedTimes)) + newRating) / float64(newRatedTimes)

		_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$set": bson.M{"rating": newTotalRating, "ratedTimes": newRatedTimes},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s MangaService) RemoveMangaRating(userID, mangaID primitive.ObjectID) error {
	ctx := context.Background()

	_, err := s.neo4j.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, `
			MATCH (u:User {id: $userID})-[r:RATED]->(m:Manga {id: $mangaID})
			DELETE r`, map[string]interface{}{
			"userID":  userID.Hex(),
			"mangaID": mangaID.Hex(),
		})
		return nil, err
	})
	if err != nil {
		return err
	}

	var user models.User
	err = s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	var ratingToRemove float64
	for _, rating := range user.Ratings {
		if rating.MangaID == mangaID.Hex() {
			ratingToRemove = rating.Score
			break
		}
	}

	_, err = s.users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$pull": bson.M{"ratings": bson.M{"mangaId": mangaID.Hex()}},
	})
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	return s.updateMangaRatingAfterRemoval(mangaID, ratingToRemove)
}

func (s MangaService) updateMangaRatingAfterRemoval(mangaID primitive.ObjectID, ratingToRemove float64) error {
	ctx := context.Background()

	var manga models.Manga
	err := s.manga.FindOne(ctx, bson.M{"_id": mangaID}).Decode(&manga)
	if err != nil {
		return err
	}

	if manga.RatedTimes > 1 {
		newRatedTimes := manga.RatedTimes - 1
		newTotalRating := ((manga.Rating * float64(manga.RatedTimes)) - ratingToRemove) / float64(newRatedTimes)

		_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$set": bson.M{"rating": newTotalRating, "ratedTimes": newRatedTimes},
		})
	} else {

		_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, bson.M{
			"$set": bson.M{"rating": 0, "ratedTimes": 0},
		})
	}
	return err
}

func (s MangaService) updatePopularMangaCache() error {
	ctx := context.Background()

	mangas, err := s.getPopularMangaFromMongo()
	if err != nil {
		return err
	}

	data, err := json.Marshal(mangas)
	if err != nil {
		logger.Error("Error marshalling popular manga for cache: " + err.Error())
		return err
	}

	if err := s.redis.Set(ctx, "popular_manga", data, time.Hour).Err(); err != nil {
		logger.Error("Error saving popular manga to Redis: " + err.Error())
		return err
	}

	return nil
}

func (s MangaService) getPopularMangaFromMongo() ([]models.Manga, error) {
	ctx := context.Background()

	findOptions := options.Find().
		SetSort(bson.D{
			{Key: "sold", Value: -1},
			{Key: "views", Value: -1},
		}).
		SetLimit(10)

	cursor, err := s.manga.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		logger.Error("Error retrieving popular manga from mongo: " + err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var mangas []models.Manga
	for cursor.Next(ctx) {
		var manga models.Manga
		if err := cursor.Decode(&manga); err == nil {
			mangas = append(mangas, manga)
		}
	}

	return mangas, nil
}
