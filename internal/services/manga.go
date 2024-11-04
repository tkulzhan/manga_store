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
			} else {
				logger.Debug("Updating popular manga cache")
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

func (s MangaService) GetMangaByID(id string) (*models.Manga, error) {
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

	return &manga, nil
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

	userUpdate := bson.M{
		"$push": bson.M{
			"purchaseHistory": purchase,
		},
	}

	mu.Lock()
	defer mu.Unlock()

	_, err = s.users.UpdateOne(ctx, bson.M{"_id": userID}, userUpdate)
	if err != nil {
		return err
	}

	mangaUpdate := bson.M{
		"$inc": bson.M{
			"quantity": -1,
			"sold":     1,
		},
	}
	_, err = s.manga.UpdateOne(ctx, bson.M{"_id": mangaID}, mangaUpdate)
	if err != nil {
		return err
	}

	return nil
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

func (s MangaService) RateManga() error {
	return nil
}

func (s MangaService) UpdateMangaRating() error {
	return nil
}

func (s MangaService) RemoveMangaRating() error {
	return nil
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
