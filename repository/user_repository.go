package repository

import (
	"context"
	"errors"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct{}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := db.Database.Collection("users").InsertOne(ctx, user)
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := db.Database.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := db.Database.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
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
// DecrementSearchesLeft atomically decrements searchesLeft by 1, only if > 0.
// Returns an error if no document was matched (i.e. searchesLeft was already 0).
func (r *UserRepository) DecrementSearchesLeft(ctx context.Context, userID primitive.ObjectID) error {
	res, err := db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID, "subscription.searchesLeft": bson.M{"$gt": 0}},
		bson.M{"$inc": bson.M{"subscription.searchesLeft": -1}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("no searches left")
	}
	return nil
}

func (r *UserRepository) DeleteUserByID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := db.Database.Collection("users").DeleteOne(
		ctx,
		bson.M{"_id": userID},
	)
	return err
}

// GetAllUsers finds users whose email matches the query (case-insensitive regex),
// or by exact ObjectID if the query is a valid 24-char hex string.
func (r *UserRepository) GetAllUsers(ctx context.Context, query string) ([]models.User, error) {
	filter := bson.M{}
	if query != "" {
		// Try matching as ObjectID first, otherwise search by email regex
		if objID, err := primitive.ObjectIDFromHex(query); err == nil {
			filter["$or"] = []bson.M{
				{"_id": objID},
				{"email": bson.M{"$regex": query, "$options": "i"}},
			}
		} else {
			filter["email"] = bson.M{"$regex": query, "$options": "i"}
		}
	}
	opts := options.Find().SetLimit(50)
	cursor, err := db.Database.Collection("users").Find(ctx, filter, opts)
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

// LinkGoogle updates a user's email, authProvider, and googleId to link with Google.
func (r *UserRepository) LinkGoogle(ctx context.Context, userID primitive.ObjectID, email, googleID string) error {
	res, err := db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"email":        email,
			"authProvider": "GOOGLE",
			"googleId":     googleID,
		}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateSearchesLeft sets subscription.searchesLeft to the given count.
func (r *UserRepository) UpdateSearchesLeft(ctx context.Context, userID primitive.ObjectID, count int) error {
	res, err := db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"subscription.searchesLeft": count}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
