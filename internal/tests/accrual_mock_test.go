package tests

import (
	"bytes"
	"context"
	"encoding/json"
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

type TestOrderProcessor struct {
	accrualService *MockAccrualService
	orderService   *mockservice.MockOrderService
}

func NewTestOrderProcessor(accrualService *MockAccrualService, orderService *mockservice.MockOrderService) *TestOrderProcessor {
	return &TestOrderProcessor{
		accrualService: accrualService,
		orderService:   orderService,
	}
}

func (p *TestOrderProcessor) ProcessOrder(t *testing.T, orderNumber string, userID int64) {
	_, err := p.accrualService.GetOrderStatus(context.Background(), orderNumber)
	assert.NoError(t, err)

	p.orderService.ProcessOrdersBackground(context.Background())
}

func TestAccrualIntegration(t *testing.T) {
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

	accrualService := NewMockAccrualService()
	processor := NewTestOrderProcessor(accrualService, mockOrderService)

	t.Run("FullOrderProcessingCycle", func(t *testing.T) {
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

		accrualService.SetOrderStatus(orderNumber, "REGISTERED", 0)

		mockOrderService.EXPECT().
			ProcessOrdersBackground(gomock.Any()).
			Times(1)

		processor.ProcessOrder(t, orderNumber, userID)

		accrualService.SetOrderStatus(orderNumber, "PROCESSING", 0)

		mockOrderService.EXPECT().
			ProcessOrdersBackground(gomock.Any()).
			Times(1)

		processor.ProcessOrder(t, orderNumber, userID)

		accrual := 500.0
		accrualService.SetOrderStatus(orderNumber, "PROCESSED", accrual)

		expectedOrders := []model.OrderResponse{
			{
				Number:     orderNumber,
				Status:     "PROCESSED",
				Accrual:    accrual,
				UploadedAt: time.Now(),
			},
		}

		mockOrderService.EXPECT().
			GetOrdersByUserID(gomock.Any(), userID).
			Return(expectedOrders, nil)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/user/orders", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		balanceResponse := model.BalanceResponse{
			Current:   accrual,
			Withdrawn: 0,
		}

		mockBalanceService.EXPECT().
			GetBalance(gomock.Any(), userID).
			Return(balanceResponse, nil)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/user/balance", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		withdrawAmount := 100.0
		withdrawOrder := "9876543210"

		mockBalanceService.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawOrder, withdrawAmount).
			Return(nil)

		withdrawRequest := map[string]interface{}{
			"order": withdrawOrder,
			"sum":   withdrawAmount,
		}
		body, _ := json.Marshal(withdrawRequest)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		withdrawalHistory := []model.WithdrawalResponse{
			{
				Order:       withdrawOrder,
				Sum:         withdrawAmount,
				ProcessedAt: time.Now(),
			},
		}

		mockBalanceService.EXPECT().
			GetWithdrawals(gomock.Any(), userID).
			Return(withdrawalHistory, nil)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/user/withdrawals", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		updatedBalanceResponse := model.BalanceResponse{
			Current:   accrual - withdrawAmount,
			Withdrawn: withdrawAmount,
		}

		mockBalanceService.EXPECT().
			GetBalance(gomock.Any(), userID).
			Return(updatedBalanceResponse, nil)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/user/balance", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
