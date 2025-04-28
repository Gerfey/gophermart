// Package service содержит бизнес-логику приложения GopherMart.
// Этот пакет реализует основные функции системы: управление пользователями,
// обработку заказов и управление балансом пользователей.
package service

import (
	"context"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
)

// UserService интерфейс для работы с пользователями.
// Предоставляет методы для регистрации, аутентификации и проверки токенов пользователей.
type UserService interface {
	// RegisterUser регистрирует нового пользователя с указанными логином и паролем.
	// Возвращает JWT токен в случае успешной регистрации или ошибку.
	RegisterUser(ctx context.Context, login, password string) (string, error)

	// LoginUser аутентифицирует пользователя с указанными логином и паролем.
	// Возвращает JWT токен в случае успешной аутентификации или ошибку.
	LoginUser(ctx context.Context, login, password string) (string, error)

	// ParseToken проверяет JWT токен и возвращает идентификатор пользователя.
	// Возвращает ошибку, если токен недействителен или истек срок его действия.
	ParseToken(token string) (int64, error)
}

// OrderService интерфейс для работы с заказами.
// Предоставляет методы для создания, получения и обработки заказов.
type OrderService interface {
	// CreateOrder создает новый заказ для пользователя с указанным номером.
	// Возвращает HTTP-код ответа и ошибку, если она возникла.
	CreateOrder(ctx context.Context, userID int64, number string) (int, error)

	// GetOrdersByUserID возвращает список заказов пользователя.
	// Возвращает ошибку, если не удалось получить заказы.
	GetOrdersByUserID(ctx context.Context, userID int64) ([]model.OrderResponse, error)

	// ProcessOrdersBackground запускает фоновую обработку заказов.
	// Периодически проверяет статус заказов в системе начислений и обновляет их в базе данных.
	ProcessOrdersBackground(ctx context.Context)
}

// BalanceService интерфейс для работы с балансом пользователей.
// Предоставляет методы для получения баланса, списания средств и получения истории списаний.
type BalanceService interface {
	// GetBalance возвращает текущий баланс пользователя.
	// Возвращает ошибку, если не удалось получить баланс.
	GetBalance(ctx context.Context, userID int64) (model.BalanceResponse, error)

	// Withdraw списывает указанную сумму с баланса пользователя на указанный заказ.
	// Возвращает ошибку, если не удалось выполнить списание.
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error

	// GetWithdrawals возвращает историю списаний пользователя.
	// Возвращает ошибку, если не удалось получить историю списаний.
	GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalResponse, error)
}

// Service структура, объединяющая все сервисы приложения.
// Предоставляет доступ к сервисам пользователей, заказов и баланса.
type Service struct {
	// Users сервис для работы с пользователями
	Users    UserService
	// Orders сервис для работы с заказами
	Orders   OrderService
	// Balances сервис для работы с балансом
	Balances BalanceService
}

// NewService создает новый экземпляр Service с инициализированными сервисами.
// Параметры:
//   - repos: репозитории для доступа к данным
//   - cfg: конфигурация приложения
//
// Возвращает:
//   - *Service: инициализированный экземпляр сервисов
func NewService(repos *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		Users:    NewUserService(repos.Users, cfg),
		Orders:   NewOrderService(repos.Orders, repos.Balances, cfg),
		Balances: NewBalanceService(repos.Balances),
	}
}
