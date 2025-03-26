package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Gerfey/gophermart/internal/tests"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

func TestRegister(t *testing.T) {
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
		creds := model.UserCredentials{
			Login:    "testuser",
			Password: "password123",
		}

		mockUserService.EXPECT().
			RegisterUser(gomock.Any(), creds.Login, creds.Password).
			Return("valid_token", nil)

		requestBody, _ := json.Marshal(creds)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Authorization"), "Bearer")
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		creds := model.UserCredentials{
			Login:    "existinguser",
			Password: "password123",
		}

		mockUserService.EXPECT().
			RegisterUser(gomock.Any(), creds.Login, creds.Password).
			Return("", fmt.Errorf("пользователь с логином %s уже существует", creds.Login))

		requestBody, _ := json.Marshal(creds)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
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
		creds := model.UserCredentials{
			Login:    "testuser",
			Password: "password123",
		}

		mockUserService.EXPECT().
			LoginUser(gomock.Any(), creds.Login, creds.Password).
			Return("valid_token", nil)

		requestBody, _ := json.Marshal(creds)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Authorization"), "Bearer")
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		creds := model.UserCredentials{
			Login:    "testuser",
			Password: "wrongpassword",
		}

		mockUserService.EXPECT().
			LoginUser(gomock.Any(), creds.Login, creds.Password).
			Return("", tests.ErrInvalidCredentials)

		requestBody, _ := json.Marshal(creds)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
