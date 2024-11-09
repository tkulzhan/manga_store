package services

import (
	"context"
	"errors"
	"fmt"
	"manga_store/internal/databases"
	"manga_store/internal/logger"
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

func (s UserService) GetUser(userID primitive.ObjectID) (*models.User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.User
    err := s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("user not found")
        }
        return nil, errors.New("failed to retrieve user from MongoDB")
    }

    return &user, nil
}


func (s UserService) GetRecsByPreferences(userID string) ([]models.Manga, error) {
	neo4jCtx := context.Background()

	// Step 1: Find genres of high-rated manga (rated > 4)
	genreQuery := `
	    MATCH (u:User {id: $userID})-[r:RATED]->(manga:Manga)
	    WHERE r.score > 4
	    WITH u, apoc.coll.flatten(collect(DISTINCT manga.genres)) AS userGenres

	    MATCH (otherManga:Manga)
	    WHERE any(genre IN otherManga.genres WHERE genre IN userGenres)
	    AND NOT (u)-[:RATED|PURCHASED]->(otherManga)
	    RETURN DISTINCT otherManga.id AS id, otherManga.genres AS genres
	    LIMIT 10
	`

	genreResult, err := s.neo4j.ExecuteRead(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(neo4jCtx, genreQuery, map[string]interface{}{
			"userID": userID,
		})
		if err != nil {
			return nil, err
		}

		var mangaIDs []string

		for res.Next(neo4jCtx) {
			record := res.Record()
			id, ok := record.Get("id")
			if ok {
				mangaIDs = append(mangaIDs, id.(string))
			}
		}

		if err = res.Err(); err != nil {
			return nil, err
		}

		return mangaIDs, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get manga IDs: %w", err)
	}

	mangaIDs, _ := genreResult.([]string)

	// Convert manga IDs to ObjectIDs for MongoDB
	mangaObjectIDs := []primitive.ObjectID{}
	for _, v := range mangaIDs {
		objectId, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manga ID to ObjectID: %w", err)
		}
		mangaObjectIDs = append(mangaObjectIDs, objectId)
	}

	// Step 2: Fetch the manga from MongoDB using the retrieved IDs
	recommendations, err := fetchMangaFromMongoDB(s, mangaObjectIDs)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

func fetchMangaFromMongoDB(s UserService, mangaObjectIDs []primitive.ObjectID) ([]models.Manga, error) {
	var recommendations []models.Manga

	if len(mangaObjectIDs) > 0 {
		filter := bson.M{"_id": bson.M{"$in": mangaObjectIDs}}

		cursor, err := s.manga.Find(context.Background(), filter)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch manga: %w", err)
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var manga models.Manga
			if err := cursor.Decode(&manga); err != nil {
				return nil, err
			}
			recommendations = append(recommendations, manga)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		logger.Debug("Retrieving popular manga")
		popular, err := NewMangaService().GetPopularManga()
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, popular...)
	}

	return recommendations, nil
}

func (s UserService) GetRecsBySimilarUsers(userID string) ([]models.Manga, error) {
	neo4jCtx := context.Background()

	// Neo4j query for collaborative filtering
	recQuery := `
        // Step 1: Find the manga rated by the target user with high ratings
        MATCH (u:User {id: $userID})-[r:RATED]->(manga:Manga)
        WHERE r.score > 4
        WITH u, manga

        // Step 2: Find other users who have rated the same manga highly
        MATCH (otherUser:User)-[otherRating:RATED]->(manga)
        WHERE otherUser <> u AND otherRating.score > 4

        // Step 3: Find other manga rated by similar users
        MATCH (otherUser)-[similarRating:RATED]->(similarManga:Manga)
        WHERE NOT (u)-[:RATED|PURCHASED]->(similarManga)
        
        // Step 4: Aggregate the recommendations and calculate the average rating
        WITH similarManga, avg(similarRating.score) AS avgRating, count(similarRating) AS ratingCount
        RETURN similarManga.id AS id, similarManga.genres AS genres, avgRating, ratingCount
        ORDER BY avgRating DESC, ratingCount DESC
        LIMIT 10
    `

	// Execute the Neo4j query
	recResult, err := s.neo4j.ExecuteRead(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(neo4jCtx, recQuery, map[string]interface{}{
			"userID": userID,
		})
		if err != nil {
			return nil, err
		}

		var mangaIDs []string
		for res.Next(neo4jCtx) {
			record := res.Record()
			id, ok := record.Get("id")
			if ok {
				mangaIDs = append(mangaIDs, id.(string))
			}
		}

		if err = res.Err(); err != nil {
			return nil, err
		}

		return mangaIDs, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get manga recommendations: %w", err)
	}

	mangaIDs, _ := recResult.([]string)

	// Convert manga IDs to ObjectIDs for MongoDB
	mangaObjectIDs := []primitive.ObjectID{}
	for _, v := range mangaIDs {
		objectId, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manga ID to ObjectID: %w", err)
		}
		mangaObjectIDs = append(mangaObjectIDs, objectId)
	}

	// Step 5: Fetch the manga from MongoDB using the retrieved IDs
	recommendations, err := fetchMangaFromMongoDB(s, mangaObjectIDs)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
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
