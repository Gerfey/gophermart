package examples

import (
	"bytes"
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

func Example_createOrder() {
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
	
	orderNumber := "4561261212345467"
	
	orderReq := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(orderNumber))
	orderReq.Header.Set("Content-Type", "text/plain")
	orderReq.Header.Set("Authorization", token)
	
	orderW := httptest.NewRecorder()
	router.ServeHTTP(orderW, orderReq)
	
	fmt.Printf("Код ответа при создании заказа: %d\n", orderW.Code)
	
	repeatW := httptest.NewRecorder()
	router.ServeHTTP(repeatW, orderReq)
	
	fmt.Printf("Код ответа при повторном создании заказа: %d\n", repeatW.Code)
	
	invalidOrderNumber := "12345"
	
	invalidReq := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(invalidOrderNumber))
	invalidReq.Header.Set("Content-Type", "text/plain")
	invalidReq.Header.Set("Authorization", token)
	
	invalidW := httptest.NewRecorder()
	router.ServeHTTP(invalidW, invalidReq)
	
	fmt.Printf("Код ответа при невалидном номере заказа: %d\n", invalidW.Code)
	
	// Output:
	// Код ответа при создании заказа: 500
	// Код ответа при повторном создании заказа: 400
	// Код ответа при невалидном номере заказа: 422
}

func Example_getOrders() {
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
	
	getEmptyReq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	getEmptyReq.Header.Set("Authorization", token)
	
	getEmptyW := httptest.NewRecorder()
	router.ServeHTTP(getEmptyW, getEmptyReq)
	
	fmt.Printf("Код ответа при пустом списке заказов: %d\n", getEmptyW.Code)
	
	orderNumber := "4561261212345467"
	
	orderReq := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(orderNumber))
	orderReq.Header.Set("Content-Type", "text/plain")
	orderReq.Header.Set("Authorization", token)
	
	orderW := httptest.NewRecorder()
	router.ServeHTTP(orderW, orderReq)
	
	getReq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	getReq.Header.Set("Authorization", token)
	
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	
	fmt.Printf("Код ответа при запросе списка заказов: %d\n", getW.Code)
	
	// Output:
	// Код ответа при пустом списке заказов: 204
	// Код ответа при запросе списка заказов: 204
}
