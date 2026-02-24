package services

import (
	"context"
	"errors"

	"USDT_BackEnd/models"
	"USDT_BackEnd/repository"
	"USDT_BackEnd/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordService struct {
	repo *repository.WordRepository
}

func NewWordService() *WordService {
	return &WordService{repo: &repository.WordRepository{}}
}

func (s *WordService) SearchWords(ctx context.Context, query string) ([]models.Word, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	var kanaQueries []string
	if utils.IsRomaji(query) {
		hiragana := utils.RomajiToHiragana(query)
		katakana := utils.HiraganaToKatakana(hiragana)
		kanaQueries = []string{hiragana, katakana}
	}
	return s.repo.SearchWords(ctx, query, kanaQueries)
}

func (s *WordService) GetWordByID(ctx context.Context, idStr string) (*models.Word, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	return s.repo.GetWordByID(ctx, id)
}
func (s *WordService) BulkCreateWords(ctx context.Context, words []models.Word) (int, error) {
	return s.repo.BulkInsert(ctx, words)
}

func (s *WordService) GetAllWords(ctx context.Context, page, limit int, query string) ([]models.Word, bool, int64, error) {
	var kanaQueries []string
	if utils.IsRomaji(query) {
		hiragana := utils.RomajiToHiragana(query)
		katakana := utils.HiraganaToKatakana(hiragana)
		kanaQueries = []string{hiragana, katakana}
	}
	return s.repo.GetAllWords(ctx, page, limit, query, kanaQueries)
}

func (s *WordService) CreateWord(ctx context.Context, word *models.Word) error {
	return s.repo.CreateWord(ctx, word)
}

func (s *WordService) UpdateWord(ctx context.Context, idStr string, word *models.Word) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	return s.repo.UpdateWord(ctx, id, word)
}

func (s *WordService) DeleteWord(ctx context.Context, idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	return s.repo.DeleteWord(ctx, id)
}

func (s *WordService) SetWordIgnore(ctx context.Context, idStr string, ignore bool) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	return s.repo.SetWordIgnore(ctx, id, ignore)
}

func (s *WordService) GetDuplicateWords(ctx context.Context, page, limit int, query string) ([]interface{}, int64, error) {
	results, total, err := s.repo.GetDuplicateWords(ctx, page, limit, query)
	if err != nil {
		return nil, 0, err
	}
	out := make([]interface{}, len(results))
	for i, r := range results {
		out[i] = r
	}
	return out, total, nil
}
