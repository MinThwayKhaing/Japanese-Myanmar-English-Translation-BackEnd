package services

import (
	"context"
	"errors"

	"USDT_BackEnd/models"
	"USDT_BackEnd/repository"

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
	return s.repo.SearchWords(ctx, query)
}

func (s *WordService) GetWordByID(ctx context.Context, idStr string) (*models.Word, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	return s.repo.GetWordByID(ctx, id)
}
func (s *WordService) BulkCreateWords(ctx context.Context, words []models.Word) error {
	return s.repo.BulkInsert(ctx, words)
}

func (s *WordService) GetAllWords(ctx context.Context, page, limit int, query string) ([]models.Word, bool, int64, error) {
	return s.repo.GetAllWords(ctx, page, limit, query)
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
