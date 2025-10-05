package services

import (
	"be/internal/models"
	"be/internal/repositories"
	"context"
)

type OrderService interface {
	Create(ctx context.Context, productID int64, quantity int, buyerID string) (*models.Order, error)
	Get(ctx context.Context, id int64) (*models.Order, error)
}

type orderService struct {
	repo repositories.OrderRepository
}

func NewOrderService(repo repositories.OrderRepository) *orderService {
	return &orderService{repo: repo}
}

func (s *orderService) Create(ctx context.Context, productID int64, quantity int, buyerID string) (*models.Order, error) {
	return s.repo.CreateOrderWithStock(ctx, productID, quantity, buyerID)
}

func (s *orderService) Get(ctx context.Context, id int64) (*models.Order, error) {
	return s.repo.GetByID(ctx, id)
}
