package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Gerfey/gophermart/internal/handler"
	"github.com/Gerfey/gophermart/internal/service"
	mockservice "github.com/Gerfey/gophermart/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func TestRegisterUser(t *testing.T) {
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

	t.Run("SuccessfulRegistration", func(t *testing.T) {
		mockUserService.EXPECT().
			RegisterUser(gomock.Any(), "testuser", "password123").
			Return("token123", nil)

		reqBody := map[string]string{
			"login":    "testuser",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Bearer token123", w.Header().Get("Authorization"))
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		mockUserService.EXPECT().
			RegisterUser(gomock.Any(), "existinguser", "password123").
			Return("", ErrUserAlreadyExists)

		reqBody := map[string]string{
			"login":    "existinguser",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginUser(t *testing.T) {
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

	t.Run("SuccessfulLogin", func(t *testing.T) {
		mockUserService.EXPECT().
			LoginUser(gomock.Any(), "testuser", "password123").
			Return("token123", nil)

		reqBody := map[string]string{
			"login":    "testuser",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Bearer token123", w.Header().Get("Authorization"))
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		mockUserService.EXPECT().
			LoginUser(gomock.Any(), "testuser", "wrongpassword").
			Return("", ErrInvalidCredentials)

		reqBody := map[string]string{
			"login":    "testuser",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
