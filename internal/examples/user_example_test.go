package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/handler"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
	"github.com/Gerfey/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

func ExampleRegisterUser() {
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
	
	fmt.Printf("Код ответа: %d\n", w.Code)
	fmt.Printf("Токен получен: %t\n", w.Header().Get("Authorization") != "")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	fmt.Printf("Код ответа при повторной регистрации: %d\n", w.Code)
	
	// Output:
	// Код ответа: 200
	// Токен получен: true
	// Код ответа при повторной регистрации: 400
}

func ExampleLoginUser() {
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
	
	loginReq := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
	loginReq.Header.Set("Content-Type", "application/json")
	
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	
	fmt.Printf("Код ответа при входе: %d\n", loginW.Code)
	fmt.Printf("Токен получен при входе: %t\n", loginW.Header().Get("Authorization") != "")
	
	wrongCredentials := model.UserCredentials{
		Login:    "testuser",
		Password: "wrongpassword",
	}
	
	wrongBody, _ := json.Marshal(wrongCredentials)
	wrongReq := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(wrongBody))
	wrongReq.Header.Set("Content-Type", "application/json")
	
	wrongW := httptest.NewRecorder()
	router.ServeHTTP(wrongW, wrongReq)
	
	fmt.Printf("Код ответа при неверном пароле: %d\n", wrongW.Code)
	
	// Output:
	// Код ответа при входе: 200
	// Токен получен при входе: true
	// Код ответа при неверном пароле: 401
}
