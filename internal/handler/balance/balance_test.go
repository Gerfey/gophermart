package balance_test

import (
	"bytes"
	"encoding/json"
	"github.com/Gerfey/gophermart/internal/tests"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestMain(m *testing.M) {
	tests.SetupTestLogging()
	os.Exit(m.Run())
}

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

	t.Run("SuccessfulBalanceRetrieval", func(t *testing.T) {
		balance := model.BalanceResponse{
			Current:   100.5,
			Withdrawn: 50.25,
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
		assert.Equal(t, balance.Current, response.Current)
		assert.Equal(t, balance.Withdrawn, response.Withdrawn)
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

	t.Run("SuccessfulWithdrawal", func(t *testing.T) {
		withdrawRequest := model.WithdrawRequest{
			Order: "1234567890",
			Sum:   50.0,
		}

		mockBalanceService.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawRequest.Order, withdrawRequest.Sum).
			Return(nil)

		requestBody, _ := json.Marshal(withdrawRequest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
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

	t.Run("SuccessfulWithdrawalsRetrieval", func(t *testing.T) {
		withdrawals := []model.WithdrawalResponse{
			{
				Order:       "1234567890",
				Sum:         50.0,
				ProcessedAt: time.Now(),
			},
			{
				Order:       "0987654321",
				Sum:         25.5,
				ProcessedAt: time.Now().Add(-24 * time.Hour),
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
		assert.Equal(t, len(withdrawals), len(response))
	})
}
