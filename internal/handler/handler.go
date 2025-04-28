// Package handler содержит HTTP обработчики для всех эндпоинтов API сервиса GopherMart.
// Обработчики используют сервисный слой для выполнения бизнес-логики и возвращают
// результаты клиентам в формате JSON.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Gerfey/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Handler структура, содержащая все HTTP обработчики для API.
// Использует сервисный слой для выполнения бизнес-логики.
type Handler struct {
	services *service.Service
}

// NewHandler создает новый экземпляр Handler с указанными сервисами.
// Параметры:
//   - services: экземпляр сервисного слоя, содержащий бизнес-логику
//
// Возвращает:
//   - *Handler: новый экземпляр обработчика
func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

// InitRoutes инициализирует все маршруты API и возвращает настроенный роутер.
// Настраивает следующие эндпоинты:
//   - POST /api/user/register - регистрация нового пользователя
//   - POST /api/user/login - аутентификация пользователя
//   - POST /api/user/orders - загрузка нового заказа (требует аутентификации)
//   - GET /api/user/orders - получение списка заказов (требует аутентификации)
//   - GET /api/user/balance - получение текущего баланса (требует аутентификации)
//   - POST /api/user/balance/withdraw - списание средств (требует аутентификации)
//   - GET /api/user/withdrawals - история списаний (требует аутентификации)
//
// Возвращает:
//   - *gin.Engine: настроенный роутер с зарегистрированными обработчиками
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("/register", h.registerUser)
			user.POST("/login", h.loginUser)

			authenticated := user.Group("/", h.userIdentity)
			{
				authenticated.POST("/orders", h.createOrder)
				authenticated.GET("/orders", h.getOrders)

				authenticated.GET("/balance", h.getBalance)
				authenticated.POST("/balance/withdraw", h.withdrawFromBalance)
				authenticated.GET("/withdrawals", h.getWithdrawals)
			}
		}
	}

	return router
}

// Server структура, представляющая HTTP-сервер приложения.
type Server struct {
	httpServer *http.Server
}

// NewServer создает новый экземпляр HTTP-сервера с указанным адресом и обработчиком.
// Параметры:
//   - addr: адрес, на котором будет запущен сервер (например, ":8080")
//   - handler: HTTP обработчик запросов (обычно gin.Engine)
//
// Возвращает:
//   - *Server: настроенный экземпляр сервера
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

// Run запускает HTTP-сервер и блокирует выполнение до завершения работы сервера.
// Возвращает ошибку, если сервер не удалось запустить или произошла ошибка во время работы.
//
// Возвращает:
//   - error: ошибка запуска или работы сервера
func (s *Server) Run() error {
	log.Infof("Запуск HTTP-сервера на %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown выполняет корректное завершение работы HTTP-сервера.
// Ожидает завершения всех активных соединений или истечения времени контекста.
//
// Параметры:
//   - ctx: контекст с таймаутом для завершения работы
//
// Возвращает:
//   - error: ошибка при завершении работы сервера
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
