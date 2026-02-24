package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WordRepository struct{}

func (r *WordRepository) SearchWords(ctx context.Context, query string, kanaQueries []string) ([]models.Word, error) {
	collection := db.Database.Collection("words")

	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	// Escape regex metacharacters so special chars in user input don't break the pattern
	escaped := regexp.QuoteMeta(q)

	// Use contains match (not prefix-only) so Hiragana, Katakana, Kanji,
	// Myanmar, English, and Romaji all match regardless of position in the field.
	orClauses := []bson.M{
		{"english": bson.M{"$regex": escaped, "$options": "i"}},
		{"japanese": bson.M{"$regex": escaped, "$options": "i"}},
		{"myanmar": bson.M{"$regex": escaped, "$options": "i"}},
		{"subTerm": bson.M{"$regex": escaped, "$options": "i"}},
	}

	// Add hiragana and katakana variants to Japanese field searches.
	for _, kana := range kanaQueries {
		if kana != "" && kana != q {
			esc := regexp.QuoteMeta(kana)
			orClauses = append(orClauses,
				bson.M{"japanese": bson.M{"$regex": esc, "$options": "i"}},
				bson.M{"subTerm": bson.M{"$regex": esc, "$options": "i"}},
			)
		}
	}

	filter := bson.M{"$or": orClauses}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().SetLimit(100)
	cursor, err := collection.Find(ctx, filter, opts)
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
	if err != nil {
		return nil, err
	}
	return &word, nil
}
func (r *WordRepository) BulkInsert(ctx context.Context, words []models.Word) (int, error) {
	collection := db.Database.Collection("words")

	if len(words) == 0 {
		return 0, nil
	}

	const batchSize = 5000
	now := time.Now()
	totalInserted := 0

	for i := 0; i < len(words); i += batchSize {
		end := i + batchSize
		if end > len(words) {
			end = len(words)
		}

		batch := words[i:end]
		docs := make([]interface{}, len(batch))
		for j, w := range batch {
			w.CreatedAt = now
			w.UpdatedAt = now
			docs[j] = w
		}

		_, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return totalInserted, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}
		totalInserted += len(batch)
	}

	return totalInserted, nil
}

func (r *WordRepository) GetAllWords(ctx context.Context, page, limit int, query string, kanaQueries []string) ([]models.Word, bool, int64, error) {
	collection := db.Database.Collection("words")
	filter := bson.M{}
	if query != "" {
		// When the query was romaji, use regex so hiragana/katakana variants are
		// searched across Japanese fields. Otherwise use the full-text index.
		if len(kanaQueries) > 0 {
			escapedQ := regexp.QuoteMeta(query)
			orClauses := []bson.M{
				{"english": bson.M{"$regex": escapedQ, "$options": "i"}},
				{"japanese": bson.M{"$regex": escapedQ, "$options": "i"}},
				{"myanmar": bson.M{"$regex": escapedQ, "$options": "i"}},
				{"subTerm": bson.M{"$regex": escapedQ, "$options": "i"}},
			}
			for _, kana := range kanaQueries {
				if kana != "" && kana != query {
					esc := regexp.QuoteMeta(kana)
					orClauses = append(orClauses,
						bson.M{"japanese": bson.M{"$regex": esc, "$options": "i"}},
						bson.M{"subTerm": bson.M{"$regex": esc, "$options": "i"}},
					)
				}
			}
			filter = bson.M{"$or": orClauses}
		} else {
			filter = bson.M{"$text": bson.M{"$search": query}}
		}
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

func (r *WordRepository) SetWordIgnore(ctx context.Context, id primitive.ObjectID, ignore bool) error {
	_, err := db.Database.Collection("words").UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"ignore": ignore, "updatedAt": time.Now()}},
	)
	return err
}

// GetDuplicateWords finds words that share the same japanese+subTerm, with pagination and optional search.
// Words with ignore=true are excluded from duplicate detection.
func (r *WordRepository) GetDuplicateWords(ctx context.Context, page, limit int, query string) ([]bson.M, int64, error) {
	collection := db.Database.Collection("words")

	// Base match: exclude ignored words
	matchStage := bson.M{"ignore": bson.M{"$ne": true}}

	// If search query provided, filter by japanese/english/myanmar/subTerm
	if query != "" {
		matchStage["$or"] = []bson.M{
			{"japanese": bson.M{"$regex": query, "$options": "i"}},
			{"english": bson.M{"$regex": query, "$options": "i"}},
			{"myanmar": bson.M{"$regex": query, "$options": "i"}},
			{"subTerm": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	// Aggregation: filter, normalize nulls, case-insensitive group by english+japanese+subTerm
	basePipeline := []bson.M{
		{"$match": matchStage},
		{"$addFields": bson.M{
			"english":  bson.M{"$ifNull": bson.A{"$english", ""}},
			"japanese": bson.M{"$ifNull": bson.A{"$japanese", ""}},
			"subTerm":  bson.M{"$ifNull": bson.A{"$subTerm", ""}},
			"myanmar":  bson.M{"$ifNull": bson.A{"$myanmar", ""}},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"english":  bson.M{"$toLower": "$english"},
				"japanese": "$japanese",
				"subTerm":  "$subTerm",
			},
			"count":         bson.M{"$sum": 1},
			"latestCreated": bson.M{"$max": "$createdAt"},
			"words": bson.M{"$push": bson.M{
				"_id":       "$_id",
				"english":   "$english",
				"japanese":  "$japanese",
				"myanmar":   "$myanmar",
				"subTerm":   "$subTerm",
				"imageUrl":  "$imageUrl",
				"ignore":    "$ignore",
				"createdAt": "$createdAt",
			}},
		}},
		{"$match": bson.M{"count": bson.M{"$gt": 1}}},
		{"$sort": bson.M{"latestCreated": -1}},
	}

	// Count total duplicate groups (copy slice to avoid mutation)
	countPipeline := make([]bson.M, len(basePipeline)+1)
	copy(countPipeline, basePipeline)
	countPipeline[len(basePipeline)] = bson.M{"$count": "total"}

	countCursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(ctx)
	var countResult []bson.M
	if err := countCursor.All(ctx, &countResult); err != nil {
		return nil, 0, err
	}
	var totalGroups int64
	if len(countResult) > 0 {
		if v, ok := countResult[0]["total"].(int32); ok {
			totalGroups = int64(v)
		} else if v, ok := countResult[0]["total"].(int64); ok {
			totalGroups = v
		}
	}

	// Paginated results (copy slice to avoid mutation)
	paginatedPipeline := make([]bson.M, len(basePipeline)+2)
	copy(paginatedPipeline, basePipeline)
	paginatedPipeline[len(basePipeline)] = bson.M{"$skip": (page - 1) * limit}
	paginatedPipeline[len(basePipeline)+1] = bson.M{"$limit": limit}

	cursor, err := collection.Aggregate(ctx, paginatedPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, totalGroups, nil
}
