package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/handler"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
	"github.com/Gerfey/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

func Example_getBalance() {
	gin.SetMode(gin.ReleaseMode)
	
	cfg := &config.Config{
		JWTSigningKey: "test-secret-key",
	}
	
	repos := repository.NewRepositoriesForTests()
	
	services := service.NewService(repos, cfg)
	
	h := handler.NewHandler(services)
	
	router := h.InitRoutes()
	
	credentials := model.UserCredentials{
		Login:    "testuser",
		Password: "password123",
	}
	
	body, _ := json.Marshal(credentials)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	token := w.Header().Get("Authorization")
	
	balanceReq := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	balanceReq.Header.Set("Authorization", token)
	
	balanceW := httptest.NewRecorder()
	router.ServeHTTP(balanceW, balanceReq)
	
	fmt.Printf("Код ответа при запросе баланса: %d\n", balanceW.Code)
	
	var balance model.BalanceResponse
	_ = json.Unmarshal(balanceW.Body.Bytes(), &balance)
	
	fmt.Printf("Текущий баланс: %.2f\n", balance.Current)
	fmt.Printf("Сумма списаний: %.2f\n", balance.Withdrawn)
	
	orderNumber := "4561261212345467"
	
	orderReq := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(orderNumber))
	orderReq.Header.Set("Content-Type", "text/plain")
	orderReq.Header.Set("Authorization", token)
	
	orderW := httptest.NewRecorder()
	router.ServeHTTP(orderW, orderReq)
	
	userID := int64(1)
	
	_ = repos.Balances.AddAccrual(context.Background(), userID, 500.0)
	
	balanceReq2 := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	balanceReq2.Header.Set("Authorization", token)
	
	balanceW2 := httptest.NewRecorder()
	router.ServeHTTP(balanceW2, balanceReq2)
	
	var balance2 model.BalanceResponse
	_ = json.Unmarshal(balanceW2.Body.Bytes(), &balance2)
	
	fmt.Printf("Баланс после начисления: %.2f\n", balance2.Current)
	
	// Output:
	// Код ответа при запросе баланса: 200
	// Текущий баланс: 0.00
	// Сумма списаний: 0.00
	// Баланс после начисления: 500.00
}

func Example_withdrawFromBalance() {
	gin.SetMode(gin.ReleaseMode)
	
	cfg := &config.Config{
		JWTSigningKey: "test-secret-key",
	}
	
	repos := repository.NewRepositoriesForTests()
	
	services := service.NewService(repos, cfg)
	
	h := handler.NewHandler(services)
	
	router := h.InitRoutes()
	
	credentials := model.UserCredentials{
		Login:    "testuser",
		Password: "password123",
	}
	
	body, _ := json.Marshal(credentials)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	token := w.Header().Get("Authorization")
	
	userID := int64(1)
	
	_ = repos.Balances.AddAccrual(context.Background(), userID, 1000.0)
	
	withdrawRequest := model.WithdrawRequest{
		Order: "4561261212345467",
		Sum:   500.0,
	}
	
	withdrawBody, _ := json.Marshal(withdrawRequest)
	withdrawReq := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(withdrawBody))
	withdrawReq.Header.Set("Content-Type", "application/json")
	withdrawReq.Header.Set("Authorization", token)
	
	withdrawW := httptest.NewRecorder()
	router.ServeHTTP(withdrawW, withdrawReq)
	
	fmt.Printf("Код ответа при списании средств: %d\n", withdrawW.Code)
	
	balanceReq := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	balanceReq.Header.Set("Authorization", token)
	
	balanceW := httptest.NewRecorder()
	router.ServeHTTP(balanceW, balanceReq)
	
	var balance model.BalanceResponse
	_ = json.Unmarshal(balanceW.Body.Bytes(), &balance)
	
	fmt.Printf("Баланс после списания: %.2f\n", balance.Current)
	fmt.Printf("Сумма списаний: %.2f\n", balance.Withdrawn)
	
	largeWithdrawRequest := model.WithdrawRequest{
		Order: "4561261212345467",
		Sum:   1000.0,
	}
	
	largeWithdrawBody, _ := json.Marshal(largeWithdrawRequest)
	largeWithdrawReq := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(largeWithdrawBody))
	largeWithdrawReq.Header.Set("Content-Type", "application/json")
	largeWithdrawReq.Header.Set("Authorization", token)
	
	largeWithdrawW := httptest.NewRecorder()
	router.ServeHTTP(largeWithdrawW, largeWithdrawReq)
	
	fmt.Printf("Код ответа при недостаточном балансе: %d\n", largeWithdrawW.Code)
	
	// Output:
	// Код ответа при списании средств: 200
	// Баланс после списания: 500.00
	// Сумма списаний: 500.00
	// Код ответа при недостаточном балансе: 402
}

func Example_getWithdrawals() {
	gin.SetMode(gin.ReleaseMode)
	
	cfg := &config.Config{
		JWTSigningKey: "test-secret-key",
	}
	
	repos := repository.NewRepositoriesForTests()
	
	services := service.NewService(repos, cfg)
	
	h := handler.NewHandler(services)
	
	router := h.InitRoutes()
	
	credentials := model.UserCredentials{
		Login:    "testuser",
		Password: "password123",
	}
	
	body, _ := json.Marshal(credentials)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	token := w.Header().Get("Authorization")
	
	withdrawalsReq := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	withdrawalsReq.Header.Set("Authorization", token)
	
	withdrawalsW := httptest.NewRecorder()
	router.ServeHTTP(withdrawalsW, withdrawalsReq)
	
	fmt.Printf("Код ответа при пустой истории списаний: %d\n", withdrawalsW.Code)
	
	userID := int64(1)
	orderNumber := "4561261212345467"
	_ = repos.Balances.AddAccrual(context.Background(), userID, 1000.0)
	
	withdrawRequest := model.WithdrawRequest{
		Order: orderNumber,
		Sum:   500.0,
	}
	
	withdrawBody, _ := json.Marshal(withdrawRequest)
	withdrawReq := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(withdrawBody))
	withdrawReq.Header.Set("Content-Type", "application/json")
	withdrawReq.Header.Set("Authorization", token)
	
	withdrawW := httptest.NewRecorder()
	router.ServeHTTP(withdrawW, withdrawReq)
	
	withdrawalsReq2 := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	withdrawalsReq2.Header.Set("Authorization", token)
	
	withdrawalsW2 := httptest.NewRecorder()
	router.ServeHTTP(withdrawalsW2, withdrawalsReq2)
	
	fmt.Printf("Код ответа после выполнения списания: %d\n", withdrawalsW2.Code)
	
	var withdrawals []model.WithdrawalResponse
	_ = json.Unmarshal(withdrawalsW2.Body.Bytes(), &withdrawals)
	
	fmt.Printf("Количество записей в истории списаний: %d\n", len(withdrawals))
	fmt.Printf("Сумма первого списания: %.2f\n", withdrawals[0].Sum)
	fmt.Printf("Номер заказа первого списания: %s\n", withdrawals[0].Order)
	
	// Output:
	// Код ответа при пустой истории списаний: 204
	// Код ответа после выполнения списания: 200
	// Количество записей в истории списаний: 1
	// Сумма первого списания: 500.00
	// Номер заказа первого списания: 4561261212345467
}
