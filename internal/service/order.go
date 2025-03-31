package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/errors"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
	log "github.com/sirupsen/logrus"
)

const (
	ErrOrderCreated      = 200
	ErrOrderAccepted     = 202
	ErrOrderNotValid     = 422
	ErrOrderRegisteredBy = 409

	defaultCheckInterval = 10 * time.Second
)

type OrderSvc struct {
	orderRepo        repository.OrderRepository
	balanceRepo      repository.BalanceRepository
	accrualSystemURL string
	checkInterval    time.Duration
}

func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository, cfg *config.Config) *OrderSvc {
	return &OrderSvc{
		orderRepo:        orderRepo,
		balanceRepo:      balanceRepo,
		accrualSystemURL: cfg.AccrualSystemAddress,
		checkInterval:    defaultCheckInterval,
	}
}

func (s *OrderSvc) CreateOrder(ctx context.Context, userID int64, number string) (int, error) {
	if !isValidLuhnNumber(number) {
		return ErrOrderNotValid, fmt.Errorf("%w", errors.ErrInvalidLuhn)
	}

	existingOrder, err := s.orderRepo.GetOrderByNumber(ctx, number)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ошибка проверки существования заказа: %w", err)
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return ErrOrderCreated, nil
		}
		return ErrOrderRegisteredBy, fmt.Errorf("%w", errors.ErrOrderAlreadyExists)
	}

	_, err = s.orderRepo.CreateOrder(ctx, userID, number)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ошибка создания заказа: %w", err)
	}

	return ErrOrderAccepted, nil
}

func (s *OrderSvc) GetOrdersByUserID(ctx context.Context, userID int64) ([]model.OrderResponse, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов: %w", err)
	}

	response := make([]model.OrderResponse, 0, len(orders))
	for _, order := range orders {
		resp := model.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		if order.Status == model.OrderStatusProcessed {
			resp.Accrual = order.Accrual
		}

		response = append(response, resp)
	}

	return response, nil
}

func (s *OrderSvc) ProcessOrdersBackground(ctx context.Context) {
	log.Info("Запуск фоновой обработки заказов")

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Остановка фоновой обработки заказов")
			return
		case <-ticker.C:
			s.processOrders(ctx)
		}
	}
}

func (s *OrderSvc) processOrders(ctx context.Context) {
	orders, err := s.orderRepo.GetNewOrProcessingOrders(ctx)
	if err != nil {
		log.Errorf("Ошибка получения заказов для проверки: %s", err.Error())
		return
	}

	for _, order := range orders {
		s.checkOrderStatus(ctx, order)
	}
}

func (s *OrderSvc) checkOrderStatus(ctx context.Context, order *model.Order) {
	baseURL, err := url.Parse(s.accrualSystemURL)
	if err != nil {
		log.Errorf("Ошибка парсинга URL системы начислений: %s", err.Error())
		return
	}

	baseURL.Path = fmt.Sprintf("/api/orders/%s", order.Number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		log.Errorf("Ошибка создания HTTP-запроса: %s", err.Error())
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Ошибка выполнения HTTP-запроса к системе начислений: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp model.AccrualResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Ошибка чтения тела ответа: %s", err.Error())
			return
		}

		if err := json.Unmarshal(body, &accrualResp); err != nil {
			log.Errorf("Ошибка разбора JSON-ответа: %s", err.Error())
			return
		}

		switch accrualResp.Status {
		case model.AccrualStatusRegistered, model.AccrualStatusProcessing:
			var newStatus model.OrderStatus
			if accrualResp.Status == model.AccrualStatusRegistered {
				newStatus = model.OrderStatusNew
			} else {
				newStatus = model.OrderStatusProcessing
			}

			if order.Status != newStatus {
				if err := s.orderRepo.UpdateOrderStatus(ctx, order.ID, newStatus); err != nil {
					log.Errorf("Ошибка обновления статуса заказа: %s", err.Error())
				}
			}
		case model.AccrualStatusInvalid:
			if err := s.orderRepo.UpdateOrderStatus(ctx, order.ID, model.OrderStatusInvalid); err != nil {
				log.Errorf("Ошибка обновления статуса заказа как невалидного: %s", err.Error())
			}
		case model.AccrualStatusProcessed:
			if err := s.orderRepo.UpdateOrderAccrual(ctx, order.ID, accrualResp.Accrual); err != nil {
				log.Errorf("Ошибка обновления начисления заказа: %s", err.Error())
				return
			}

			if err := s.balanceRepo.AddAccrual(ctx, order.UserID, accrualResp.Accrual); err != nil {
				log.Errorf("Ошибка добавления начисления к балансу: %s", err.Error())
			}
		}
	case http.StatusNoContent:
		log.Warnf("Заказ %s не зарегистрирован в системе расчета", order.Number)
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err == nil {
				log.Warnf("Превышен лимит запросов, повторная попытка через %d секунд", seconds)

				retryTimer := time.NewTimer(time.Duration(seconds) * time.Second)
				go func() {
					defer retryTimer.Stop()

					select {
					case <-ctx.Done():
						return
					case <-retryTimer.C:
						log.Infof("Повторная попытка проверки статуса заказа %s после ожидания", order.Number)
						s.checkOrderStatus(ctx, order)
					}
				}()
			} else {
				log.Warnf("Не удалось распарсить заголовок Retry-After: %s", err.Error())
			}
		} else {
			log.Warn("Получен статус TooManyRequests без заголовка Retry-After")
		}
	default:
		log.Errorf("Неожиданный ответ от системы начислений: %d", resp.StatusCode)
	}
}

func isValidLuhnNumber(number string) bool {
	number = strings.TrimSpace(number)

	for _, c := range number {
		if c < '0' || c > '9' {
			return false
		}
	}

	var sum int
	var alternate bool

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit - 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
