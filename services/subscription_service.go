// services/subscription_service.go
package services

import (
	"context"

	"USDT_BackEnd/models"
	"USDT_BackEnd/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		repo: &repository.SubscriptionRepository{},
	}
}

func (s *SubscriptionService) Create(ctx context.Context, sub models.Subscription) error {
	return s.repo.Insert(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetOne(ctx context.Context, id primitive.ObjectID) (*models.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) GetAll(ctx context.Context) ([]models.Subscription, error) {
	return s.repo.GetAll(ctx)
}

func (s *SubscriptionService) GetPaginated(
	ctx context.Context,
	search string,
	page int,
	limit int,
) ([]models.Subscription, int64, error) {
	return s.repo.GetPaginated(ctx, search, page, limit)
}
