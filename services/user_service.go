package services

import (
	"context"
	"errors"
	"log"
	"time"

	"USDT_BackEnd/config"
	"USDT_BackEnd/db"
	"USDT_BackEnd/models"
	"USDT_BackEnd/repository"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo   *repository.UserRepository
	config *config.Config
}

func NewUserService(cfg *config.Config) *UserService {
	return &UserService{repo: &repository.UserRepository{}, config: cfg}
}

// Register new user
func (s *UserService) Register(ctx context.Context, email, password string) (*models.User, error) {
	log.Println("[DEBUG] Register called for email:", email)

	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil && existing != nil {
		log.Println("[DEBUG] Email already exists:", email)
		return nil, errors.New("email already exists")
	}

	if err != nil && err.Error() != "mongo: no documents in result" {
		log.Println("[ERROR] Error checking user email:", err)
		return nil, err
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		Email:     email,
		Password:  string(hashed),
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		log.Println("[ERROR] Failed to create user:", err)
		return nil, err
	}

	log.Println("[DEBUG] User registered successfully:", email)
	return user, nil
}

// Login
// Login authenticates a user and returns JWT token
func (s *UserService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	log.Println("[DEBUG] Login called for email:", email)

	// 1️⃣ Fetch user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Println("[DEBUG] User not found:", email)
			return "", nil, errors.New("invalid credentials")
		}
		log.Println("[ERROR] Failed to fetch user:", err)
		return "", nil, err
	}

	// 2️⃣ Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Println("[DEBUG] Password mismatch for user:", email)
		return "", nil, errors.New("invalid credentials")
	}

	// 3️⃣ Generate JWT token
	tokenClaims := jwt.MapClaims{
		"user_id": user.ID.Hex(), // ensures correct MongoDB ObjectID
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.config.TokenExpiry)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		log.Println("[ERROR] Failed to sign JWT:", err)
		return "", nil, err
	}

	log.Println("[DEBUG] Login successful for email:", email)
	return tokenString, user, nil
}

// Change password
func (s *UserService) ChangePassword(ctx context.Context, userID primitive.ObjectID, current, new string) error {
	log.Println("[DEBUG] ChangePassword called for userID:", userID.Hex())

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		log.Println("[ERROR] User not found:", err)
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(current)); err != nil {
		log.Println("[DEBUG] Current password incorrect for userID:", userID.Hex())
		return errors.New("current password incorrect")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(new), bcrypt.DefaultCost)
	if err := s.repo.UpdatePassword(ctx, userID, string(hashed)); err != nil {
		log.Println("[ERROR] Failed to update password:", err)
		return err
	}

	log.Println("[DEBUG] Password changed successfully for userID:", userID.Hex())
	return nil
}

// Get favorites
func (s *UserService) GetFavoritesByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Word, error) {
	log.Println("[DEBUG] GetFavoritesByUserID called for userID:", userID.Hex())

	var user models.User
	err := db.Database.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		log.Println("[ERROR] Failed to fetch user:", err)
		return nil, err
	}

	if len(user.Favorites) == 0 {
		log.Println("[DEBUG] No favorites found for userID:", userID.Hex())
		return []models.Word{}, nil
	}

	cursor, err := db.Database.Collection("words").Find(ctx, bson.M{"_id": bson.M{"$in": user.Favorites}})
	if err != nil {
		log.Println("[ERROR] Failed to fetch favorite words:", err)
		return nil, err
	}

	var words []models.Word
	if err := cursor.All(ctx, &words); err != nil {
		log.Println("[ERROR] Failed to decode favorite words:", err)
		return nil, err
	}

	log.Println("[DEBUG] Favorites fetched for userID:", userID.Hex(), "count:", len(words))
	return words, nil
}

// Save favorite
func (s *UserService) SaveFavorite(ctx context.Context, userID, wordID primitive.ObjectID) error {
	log.Println("[DEBUG] SaveFavorite called userID:", userID.Hex(), "wordID:", wordID.Hex())
	return s.repo.SaveFavorite(ctx, userID, wordID)
}

// Remove favorite
func (s *UserService) RemoveFavorite(ctx context.Context, userID, wordID primitive.ObjectID) error {
	log.Println("[DEBUG] RemoveFavorite called userID:", userID.Hex(), "wordID:", wordID.Hex())
	return s.repo.RemoveFavorite(ctx, userID, wordID)
}

// Get favorites paginated
func (s *UserService) GetFavoritesPaginated(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Word, bool, error) {
	log.Println("[DEBUG] GetFavoritesPaginated called userID:", userID.Hex(), "page:", page, "limit:", limit)
	return s.repo.GetFavoritesPaginated(ctx, userID, page, limit)
}

// ------------------- Admin: Get Subscribed Users -------------------
func (s *UserService) GetSubscribedUsers(ctx context.Context) ([]models.User, error) {
	log.Println("[DEBUG] UserService.GetSubscribedUsers called")
	users, err := s.repo.GetSubscribedUsers(ctx)
	if err != nil {
		log.Println("[ERROR] Failed to fetch subscribed users:", err)
		return nil, err
	}
	log.Println("[DEBUG] Subscribed users fetched, count:", len(users))
	return users, nil
}
func (s *UserService) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
