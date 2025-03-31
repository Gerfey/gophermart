package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Gerfey/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

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

type Server struct {
	httpServer *http.Server
}

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

func (s *Server) Run() error {
	log.Infof("Запуск HTTP-сервера на %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
