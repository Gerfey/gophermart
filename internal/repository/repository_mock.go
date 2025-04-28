package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	customerrors "github.com/Gerfey/gophermart/internal/errors"
	"github.com/Gerfey/gophermart/internal/model"
)

var (
	ErrUserExists             = errors.New("пользователь с таким логином уже существует")
	ErrUserNotFound           = errors.New("пользователь не найден")
	ErrOrderNotFound          = errors.New("заказ не найден")
	ErrOrderAlreadyExists     = errors.New("заказ уже зарегистрирован этим пользователем")
	ErrOrderBelongsToAnotherUser = errors.New("заказ уже зарегистрирован другим пользователем")
	ErrInsufficientFunds      = customerrors.ErrInsufficientFunds
)

func NewRepositoriesForTests() *Repository {
	return &Repository{
		Users:    NewUserRepoMock(),
		Orders:   NewOrderRepoMock(),
		Balances: NewBalanceRepoMock(),
	}
}

type UserRepoMock struct {
	users map[int64]*model.User
	mutex sync.RWMutex
	lastID int64
}

func NewUserRepoMock() *UserRepoMock {
	return &UserRepoMock{
		users: make(map[int64]*model.User),
		lastID: 0,
	}
}

func (r *UserRepoMock) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	for _, user := range r.users {
		if user.Login == login {
			return 0, ErrUserExists
		}
	}
	
	r.lastID++
	userID := r.lastID
	
	r.users[userID] = &model.User{
		ID:           userID,
		Login:        login,
		PasswordHash: passwordHash,
	}
	
	return userID, nil
}

func (r *UserRepoMock) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, user := range r.users {
		if user.Login == login {
			return user, nil
		}
	}
	
	return nil, ErrUserNotFound
}

func (r *UserRepoMock) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	
	return user, nil
}

type OrderRepoMock struct {
	orders map[int64]*model.Order
	mutex sync.RWMutex
	lastID int64
}

func NewOrderRepoMock() *OrderRepoMock {
	return &OrderRepoMock{
		orders: make(map[int64]*model.Order),
		lastID: 0,
	}
}

func (r *OrderRepoMock) CreateOrder(ctx context.Context, userID int64, number string) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	for _, order := range r.orders {
		if order.Number == number {
			if order.UserID == userID {
				return order.ID, ErrOrderAlreadyExists
			}
			return 0, ErrOrderBelongsToAnotherUser
		}
	}
	
	r.lastID++
	orderID := r.lastID
	
	r.orders[orderID] = &model.Order{
		ID:         orderID,
		UserID:     userID,
		Number:     number,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now(),
	}
	
	return orderID, nil
}

func (r *OrderRepoMock) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, order := range r.orders {
		if order.Number == number {
			return order, nil
		}
	}
	
	return nil, ErrOrderNotFound
}

func (r *OrderRepoMock) GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var userOrders []*model.Order
	
	for _, order := range r.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}
	
	return userOrders, nil
}

func (r *OrderRepoMock) UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	order, exists := r.orders[orderID]
	if !exists {
		return ErrOrderNotFound
	}
	
	order.Status = status
	return nil
}

func (r *OrderRepoMock) UpdateOrderAccrual(ctx context.Context, orderID int64, accrual float64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	order, exists := r.orders[orderID]
	if !exists {
		return ErrOrderNotFound
	}
	
	order.Accrual = accrual
	order.Status = model.OrderStatusProcessed
	return nil
}

func (r *OrderRepoMock) GetNewOrProcessingOrders(ctx context.Context) ([]*model.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var result []*model.Order
	
	for _, order := range r.orders {
		if order.Status == model.OrderStatusNew || order.Status == model.OrderStatusProcessing {
			result = append(result, order)
		}
	}
	
	return result, nil
}

type BalanceRepoMock struct {
	balances   map[int64]*model.Balance
	withdrawals map[int64][]*model.Withdrawal
	mutex      sync.RWMutex
	lastID     int64
}

func NewBalanceRepoMock() *BalanceRepoMock {
	return &BalanceRepoMock{
		balances:   make(map[int64]*model.Balance),
		withdrawals: make(map[int64][]*model.Withdrawal),
		lastID:     0,
	}
}

func (r *BalanceRepoMock) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	balance, exists := r.balances[userID]
	if !exists {
		return &model.Balance{
			UserID:    userID,
			Current:   0,
			Withdrawn: 0,
		}, nil
	}
	
	return balance, nil
}

func (r *BalanceRepoMock) CreateBalance(ctx context.Context, userID int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.balances[userID] = &model.Balance{
		UserID:    userID,
		Current:   0,
		Withdrawn: 0,
	}
	
	return nil
}

func (r *BalanceRepoMock) AddAccrual(ctx context.Context, userID int64, amount float64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	balance, exists := r.balances[userID]
	if !exists {
		r.balances[userID] = &model.Balance{
			UserID:    userID,
			Current:   amount,
			Withdrawn: 0,
		}
		return nil
	}
	
	balance.Current += amount
	return nil
}

func (r *BalanceRepoMock) Withdraw(ctx context.Context, userID int64, amount float64, orderNumber string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	balance, exists := r.balances[userID]
	if !exists {
		return customerrors.ErrInsufficientFunds
	}
	
	if balance.Current < amount {
		return customerrors.ErrInsufficientFunds
	}
	
	balance.Current -= amount
	balance.Withdrawn += amount
	
	withdrawal := &model.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		ProcessedAt: time.Now(),
	}
	
	r.withdrawals[userID] = append(r.withdrawals[userID], withdrawal)
	
	return nil
}

func (r *BalanceRepoMock) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	withdrawals, exists := r.withdrawals[userID]
	if !exists || len(withdrawals) == 0 {
		return nil, nil
	}
	
	return withdrawals, nil
}

func (r *BalanceRepoMock) AddPoints(userID int64, amount float64, orderNumber string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	balance, exists := r.balances[userID]
	if !exists {
		r.balances[userID] = &model.Balance{
			UserID:    userID,
			Current:   amount,
			Withdrawn: 0,
		}
		return
	}
	
	balance.Current += amount
}
