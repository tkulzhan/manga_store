package services

import (
	"context"
	"errors"
	"manga_store/internal/databases"
	"manga_store/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *mongo.Collection
}

func NewAuthService() AuthService {
	return AuthService{
		users: databases.Users(),
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
	var user models.User
	user.Email = email
	user.PasswordHash = string(passwordHash)

	_, err = s.users.InsertOne(ctx, user)
	return err
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
