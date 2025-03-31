package repository

import (
	"context"

	"github.com/Gerfey/gophermart/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID int64, number string) (int64, error)
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus) error
	UpdateOrderAccrual(ctx context.Context, orderID int64, accrual float64) error
	GetNewOrProcessingOrders(ctx context.Context) ([]*model.Order, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	CreateBalance(ctx context.Context, userID int64) error
	AddAccrual(ctx context.Context, userID int64, amount float64) error
	Withdraw(ctx context.Context, userID int64, amount float64, orderNumber string) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}

type Repository struct {
	Users    UserRepository
	Orders   OrderRepository
	Balances BalanceRepository
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Users:    NewUserRepo(db),
		Orders:   NewOrderRepo(db),
		Balances: NewBalanceRepo(db),
	}
}
