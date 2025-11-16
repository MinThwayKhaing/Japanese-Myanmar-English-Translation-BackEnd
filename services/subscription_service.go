package services

import (
	"context"

	"USDT_BackEnd/models"
	"USDT_BackEnd/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{repo: &repository.SubscriptionRepository{}}
}

func (s *SubscriptionService) GetPrices(ctx context.Context) ([]models.SubscriptionPlan, error) {
	return s.repo.GetPrices(ctx)
}

func (s *SubscriptionService) UpdatePrices(ctx context.Context, monthly, yearly float64, freeMonths int) error {
	return s.repo.UpdatePrices(ctx, monthly, yearly, freeMonths)
}
