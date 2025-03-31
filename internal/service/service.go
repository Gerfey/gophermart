package service

import (
	"context"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseToken(token string) (int64, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, number string) (int, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]model.OrderResponse, error)
	ProcessOrdersBackground(ctx context.Context)
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (model.BalanceResponse, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalResponse, error)
}

type Service struct {
	Users    UserService
	Orders   OrderService
	Balances BalanceService
}

func NewService(repos *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		Users:    NewUserService(repos.Users, cfg),
		Orders:   NewOrderService(repos.Orders, repos.Balances, cfg),
		Balances: NewBalanceService(repos.Balances),
	}
}
