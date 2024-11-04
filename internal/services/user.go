package services

import (
	"context"
	"errors"
	"manga_store/internal/databases"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	users *mongo.Collection
	neo4j neo4j.SessionWithContext
}

func NewUserService() UserService {
	return UserService{
		users: databases.Users(),
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
