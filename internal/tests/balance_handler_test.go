package tests

import (
	"bytes"
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

func TestGetBalance(t *testing.T) {
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

	t.Run("SuccessfulGetBalance", func(t *testing.T) {
		balance := model.BalanceResponse{
			Current:   500.5,
			Withdrawn: 42.0,
		}
		mockBalanceService.EXPECT().
			GetBalance(gomock.Any(), userID).
			Return(balance, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/balance", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response model.BalanceResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 500.5, response.Current)
		assert.Equal(t, 42.0, response.Withdrawn)
	})
}

func TestWithdraw(t *testing.T) {
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

	t.Run("SuccessfulWithdraw", func(t *testing.T) {
		withdrawRequest := map[string]interface{}{
			"order": "2377225624",
			"sum":   100.0,
		}

		mockBalanceService.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", 100.0).
			Return(nil)

		body, _ := json.Marshal(withdrawRequest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InsufficientFunds", func(t *testing.T) {
		withdrawRequest := map[string]interface{}{
			"order": "2377225624",
			"sum":   1000.0,
		}

		mockBalanceService.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", 1000.0).
			Return(ErrInsufficientFunds)

		body, _ := json.Marshal(withdrawRequest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("InvalidOrderNumber", func(t *testing.T) {
		withdrawRequest := map[string]interface{}{
			"order": "invalid",
			"sum":   100.0,
		}

		mockBalanceService.EXPECT().
			Withdraw(gomock.Any(), userID, "invalid", 100.0).
			Return(ErrInvalidOrderNumber)

		body, _ := json.Marshal(withdrawRequest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetWithdrawals(t *testing.T) {
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

	t.Run("SuccessfulGetWithdrawals", func(t *testing.T) {
		withdrawals := []model.WithdrawalResponse{
			{
				Order:       "2377225624",
				Sum:         100.0,
				ProcessedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Order:       "2377225625",
				Sum:         200.0,
				ProcessedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		}

		mockBalanceService.EXPECT().
			GetWithdrawals(gomock.Any(), userID).
			Return(withdrawals, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/withdrawals", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []model.WithdrawalResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "2377225624", response[0].Order)
		assert.Equal(t, 100.0, response[0].Sum)
	})

	t.Run("NoWithdrawals", func(t *testing.T) {
		mockBalanceService.EXPECT().
			GetWithdrawals(gomock.Any(), userID).
			Return([]model.WithdrawalResponse{}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/withdrawals", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
