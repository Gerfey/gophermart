package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Gerfey/gophermart/internal/handler"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/service"
	mockservice "github.com/Gerfey/gophermart/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	ErrOrderAlreadyUploadedByOtherUser = errors.New("order already uploaded by other user")
	ErrInvalidOrderNumber              = errors.New("invalid order number")
	ErrOrderNotFound                   = errors.New("order not found")
	ErrInsufficientFunds               = errors.New("insufficient funds")
)

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mockservice.NewMockUserService(ctrl)
	mockOrderService := mockservice.NewMockOrderService(ctrl)
	mockBalanceService := mockservice.NewMockBalanceService(ctrl)

	services := &service.Service{
		Users:    mockUserService,
		Orders:   mockOrderService,
		Balances: mockBalanceService,
	}

	h := handler.NewHandler(services)
	router := h.InitRoutes()

	userID := int64(1)
	mockUserService.EXPECT().
		ParseToken(gomock.Any()).
		Return(userID, nil).
		AnyTimes()

	t.Run("SuccessfulOrderCreation", func(t *testing.T) {
		orderNumber := "1234567890"

		mockOrderService.EXPECT().
			CreateOrder(gomock.Any(), userID, orderNumber).
			Return(http.StatusAccepted, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("OrderAlreadyUploadedByUser", func(t *testing.T) {
		orderNumber := "1234567890"

		mockOrderService.EXPECT().
			CreateOrder(gomock.Any(), userID, orderNumber).
			Return(http.StatusOK, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("OrderAlreadyUploadedByOtherUser", func(t *testing.T) {
		orderNumber := "1234567890"

		mockOrderService.EXPECT().
			CreateOrder(gomock.Any(), userID, orderNumber).
			Return(http.StatusConflict, ErrOrderAlreadyUploadedByOtherUser)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("InvalidOrderNumber", func(t *testing.T) {
		orderNumber := "invalid"

		mockOrderService.EXPECT().
			CreateOrder(gomock.Any(), userID, orderNumber).
			Return(http.StatusUnprocessableEntity, ErrInvalidOrderNumber)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestGetOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mockservice.NewMockUserService(ctrl)
	mockOrderService := mockservice.NewMockOrderService(ctrl)
	mockBalanceService := mockservice.NewMockBalanceService(ctrl)

	services := &service.Service{
		Users:    mockUserService,
		Orders:   mockOrderService,
		Balances: mockBalanceService,
	}

	h := handler.NewHandler(services)
	router := h.InitRoutes()

	userID := int64(1)
	mockUserService.EXPECT().
		ParseToken(gomock.Any()).
		Return(userID, nil).
		AnyTimes()

	t.Run("SuccessfulGetOrders", func(t *testing.T) {
		orders := []model.OrderResponse{
			{
				Number:     "1234567890",
				Status:     "PROCESSED",
				Accrual:    100.0,
				UploadedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Number:     "0987654321",
				Status:     "PROCESSING",
				UploadedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		}

		mockOrderService.EXPECT().
			GetOrdersByUserID(gomock.Any(), userID).
			Return(orders, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/orders", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []model.OrderResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "1234567890", response[0].Number)
		assert.Equal(t, model.OrderStatus("PROCESSED"), response[0].Status)
		assert.Equal(t, 100.0, response[0].Accrual)
	})

	t.Run("NoOrders", func(t *testing.T) {
		mockOrderService.EXPECT().
			GetOrdersByUserID(gomock.Any(), userID).
			Return([]model.OrderResponse{}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/orders", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

type MockAccrualService struct {
	OrderStatuses map[string]model.AccrualResponse
}

func NewMockAccrualService() *MockAccrualService {
	return &MockAccrualService{
		OrderStatuses: make(map[string]model.AccrualResponse),
	}
}

func (m *MockAccrualService) SetOrderStatus(orderNumber string, status model.AccrualSystemStatus, accrual float64) {
	m.OrderStatuses[orderNumber] = model.AccrualResponse{
		Order:   orderNumber,
		Status:  status,
		Accrual: accrual,
	}
}

func (m *MockAccrualService) GetOrderStatus(ctx context.Context, orderNumber string) (model.AccrualResponse, error) {
	if resp, ok := m.OrderStatuses[orderNumber]; ok {
		return resp, nil
	}
	return model.AccrualResponse{}, ErrOrderNotFound
}
