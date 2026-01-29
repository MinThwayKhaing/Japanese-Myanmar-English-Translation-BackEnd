package repository

import (
	"context"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository struct{}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := db.Database.Collection("users").InsertOne(ctx, user)
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := db.Database.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return &user, err
}

func (r *UserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := db.Database.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	_, err := db.Database.Collection("users").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"password": hashedPassword}})
	return err
}

func (r *UserRepository) GetSubscribedUsers(ctx context.Context) ([]models.User, error) {
	cursor, err := db.Database.Collection("users").Find(ctx, bson.M{"subscription.status": models.StatusActive})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetFavorites(ctx context.Context, favoriteIDs []primitive.ObjectID) ([]models.Word, error) {
	var words []models.Word
	cursor, err := db.Database.Collection("words").Find(ctx, bson.M{"_id": bson.M{"$in": favoriteIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &words); err != nil {
		return nil, err
	}
	return words, nil
}

// SaveFavorite adds a word ID to user's favorites (no duplicates)
func (r *UserRepository) SaveFavorite(ctx context.Context, userID, wordID primitive.ObjectID) error {
	_, err := db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$addToSet": bson.M{"favorites": wordID}}, // prevents duplicates
	)
	return err
}

// RemoveFavorite removes a word from user's favorites
func (r *UserRepository) RemoveFavorite(ctx context.Context, userID, wordID primitive.ObjectID) error {
	_, err := db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$pull": bson.M{"favorites": wordID}},
	)
	return err
}

// GetFavoritesPaginated returns favorites with pagination
func (r *UserRepository) GetFavoritesPaginated(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Word, bool, error) {
	collection := db.Database.Collection("users")

	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, false, err
	}

	// Pagination logic
	start := (page - 1) * limit
	end := start + limit
	if start >= len(user.Favorites) {
		return []models.Word{}, false, nil
	}
	if end > len(user.Favorites) {
		end = len(user.Favorites)
	}

	pagedFavorites := user.Favorites[start:end]

	// Fetch words from "words" collection
	var words []models.Word
	cursor, err := db.Database.Collection("words").Find(ctx, bson.M{"_id": bson.M{"$in": pagedFavorites}})
	if err != nil {
		return nil, false, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &words); err != nil {
		return nil, false, err
	}

	hasMore := end < len(user.Favorites)
	return words, hasMore, nil
}
func (r *UserRepository) DeleteUserByID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := db.Database.Collection("users").DeleteOne(
		ctx,
		bson.M{"_id": userID},
	)
	return err
}
