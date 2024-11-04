package services

import (
	"context"
	"errors"
	"manga_store/internal/databases"
	"manga_store/internal/models"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *mongo.Collection
	neo4j neo4j.SessionWithContext
}

func NewAuthService() AuthService {
	return AuthService{
		users: databases.Users(),
		neo4j: databases.Neo4j(context.Background()),
	}
}

func (s AuthService) Register(email, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingUser models.User
	err := s.users.FindOne(ctx, bson.M{"email": email, "isDeleted": false}).Decode(&existingUser)
	if err == nil {
		return errors.New("user with this email already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	user := models.User{
		Email:          email,
		PasswordHash:   string(passwordHash),
		PurchaseHistory: []models.Purchase{},
	}
	result, err := s.users.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	userID := result.InsertedID.(primitive.ObjectID).Hex()

	neo4jCtx := context.Background()
	_, err = s.neo4j.ExecuteWrite(neo4jCtx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(neo4jCtx, "CREATE (u:User {id: $id, email: $email})", map[string]interface{}{
			"id":    userID,
			"email": email,
		})
		return nil, err
	})

	if err != nil {
		_, _ = s.users.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
		return errors.New("failed to create user in Neo4j, registration rolled back")
	}

	return nil
}

func (s AuthService) Login(email, password string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := s.users.FindOne(ctx, bson.M{"email": email, "isDeleted": false}).Decode(&user)
	if err != nil {
		return models.User{}, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return models.User{}, errors.New("invalid email or password")
	}

	return user, nil
}

func (s AuthService) Logout() error {
	return nil
}
