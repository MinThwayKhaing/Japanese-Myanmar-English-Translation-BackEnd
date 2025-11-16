package repository

import (
	"context"
	"time"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WordRepository struct{}

func (r *WordRepository) SearchWords(ctx context.Context, query string) ([]models.Word, error) {
	collection := db.Database.Collection("words")

	// Create case-insensitive, partial match search using regex
	filter := bson.M{
		"$or": []bson.M{
			{"english": bson.M{"$regex": query, "$options": "i"}},
			{"japanese": bson.M{"$regex": query, "$options": "i"}},
			{"myanmar": bson.M{"$regex": query, "$options": "i"}},
			{"subTerm": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var words []models.Word
	if err := cursor.All(ctx, &words); err != nil {
		return nil, err
	}

	return words, nil
}

func (r *WordRepository) GetWordByID(ctx context.Context, id primitive.ObjectID) (*models.Word, error) {
	collection := db.Database.Collection("words")
	var word models.Word
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&word)
	return &word, err
}
func (r *WordRepository) BulkInsert(ctx context.Context, words []models.Word) error {
	collection := db.Database.Collection("words")

	var docs []interface{}
	now := time.Now()
	for _, w := range words {
		w.CreatedAt = now
		w.UpdatedAt = now
		docs = append(docs, w)
	}

	if len(docs) == 0 {
		return nil
	}

	_, err := collection.InsertMany(ctx, docs)
	return err
}

func (r *WordRepository) GetAllWords(ctx context.Context, page, limit int, query string) ([]models.Word, bool, int64, error) {
	collection := db.Database.Collection("words")
	filter := bson.M{}
	if query != "" {
		filter = bson.M{"$text": bson.M{"$search": query}}
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Descending order

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, false, 0, err
	}
	defer cursor.Close(ctx)

	var words []models.Word
	if err := cursor.All(ctx, &words); err != nil {
		return nil, false, 0, err
	}

	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, false, 0, err
	}

	hasMore := int64(page*limit) < totalCount

	return words, hasMore, totalCount, nil
}

func (r *WordRepository) CreateWord(ctx context.Context, word *models.Word) error {
	word.CreatedAt = time.Now()
	word.UpdatedAt = time.Now()
	_, err := db.Database.Collection("words").InsertOne(ctx, word)
	return err
}

func (r *WordRepository) UpdateWord(ctx context.Context, id primitive.ObjectID, word *models.Word) error {
	word.UpdatedAt = time.Now()
	_, err := db.Database.Collection("words").UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": word})
	return err
}

func (r *WordRepository) DeleteWord(ctx context.Context, id primitive.ObjectID) error {
	_, err := db.Database.Collection("words").DeleteOne(ctx, bson.M{"_id": id})
	return err
}
