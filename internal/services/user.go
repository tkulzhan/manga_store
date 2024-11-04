package services

import (
	"context"
	"errors"
	"fmt"
	"manga_store/internal/databases"
	"manga_store/internal/models"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	users *mongo.Collection
	manga *mongo.Collection
	neo4j neo4j.SessionWithContext
}

func NewUserService() UserService {
	return UserService{
		users: databases.Users(),
		manga: databases.Manga(),
		neo4j: databases.Neo4j(context.Background()),
	}
}

func (s UserService) GetUserByPreferences() error {
	return nil
}

func (s UserService) GetUserBySimilarUsers() error {
	return nil
}

func (s UserService) DeleteUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"isDeleted": true}})
	if err != nil {
		return err
	}

	neo4jCtx := context.Background()
	_, err = s.neo4j.ExecuteWrite(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(neo4jCtx, `
			MATCH (u:User {id: $id})
			DETACH DELETE u
		`, map[string]interface{}{
			"id": userID.Hex(),
		})
		return nil, err
	})

	if err != nil {
		return errors.New("failed to delete user in Neo4j")
	}

	return nil
}

func (s UserService) RestoreUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Update MongoDB User
	_, err := s.users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"isDeleted": false}})
	if err != nil {
		return fmt.Errorf("failed to update user status in MongoDB: %w", err)
	}

	// Step 2: Retrieve User from MongoDB
	var user models.User
	err = s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return fmt.Errorf("failed to retrieve user from MongoDB: %w", err)
	}

	neo4jCtx := context.Background()
	_, err = s.neo4j.ExecuteWrite(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Step 3: Restore or Create User Node in Neo4j
		_, err := tx.Run(neo4jCtx, `
			MERGE (u:User {id: $userID})
			SET u.email = $email
		`, map[string]interface{}{
			"userID": userID.Hex(),
			"email":  user.Email,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create/update user node in Neo4j: %w", err)
		}

		// Step 4: Restore Ratings Relationships
		for _, rating := range user.Ratings {
			var manga models.Manga
			mangaObjectId, _ := primitive.ObjectIDFromHex(rating.MangaID)
			err = s.manga.FindOne(ctx, bson.M{"_id": mangaObjectId}).Decode(&manga)
			if err != nil {
				return nil, fmt.Errorf("failed to find manga for rating in MongoDB: %w", err)
			}

			_, err = tx.Run(neo4jCtx, `
				MERGE (u:User {id: $userID})
				MERGE (m:Manga {id: $mangaID})
				ON CREATE SET m.title = $title, m.genres = $genres
				MERGE (u)-[r:RATED]->(m)
				SET r.score = $score
			`, map[string]interface{}{
				"userID":  userID.Hex(),
				"mangaID": rating.MangaID,
				"title":   manga.Title,
				"genres":  manga.Genres,
				"score":   rating.Score,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create/update RATED relationship in Neo4j: %w", err)
			}
		}

		// Step 5: Restore Purchase Relationships
		for _, purchase := range user.PurchaseHistory {
			var manga models.Manga
			mangaObjectId, _ := primitive.ObjectIDFromHex(purchase.MangaID)
			err = s.manga.FindOne(ctx, bson.M{"_id": mangaObjectId}).Decode(&manga)
			if err != nil {
				return nil, fmt.Errorf("failed to find manga for purchase in MongoDB: %w", err)
			}

			_, err = tx.Run(neo4jCtx, `
				MERGE (u:User {id: $userID})
				MERGE (m:Manga {id: $mangaID})
				ON CREATE SET m.title = $title, m.genres = $genres
				MERGE (u)-[p:PURCHASED]->(m)
			`, map[string]interface{}{
				"userID":       userID.Hex(),
				"mangaID":      purchase.MangaID,
				"title":        manga.Title,
				"genres":       manga.Genres,
				"purchaseDate": purchase.PurchaseDate,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create/update PURCHASED relationship in Neo4j: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to restore user relationships in Neo4j: %w", err)
	}

	return nil
}
