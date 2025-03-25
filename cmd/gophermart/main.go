package main

import (
	"context"
	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/handler"
	"github.com/Gerfey/gophermart/internal/repository"
	"github.com/Gerfey/gophermart/internal/service"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %s", err.Error())
	}

	log.Info(cfg.AccrualSystemAddress)

	db, err := repository.NewPostgresDB(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %s", err.Error())
	}

	repos := repository.NewRepository(db)

	services := service.NewService(repos, cfg)

	handlers := handler.NewHandler(services)

	server := handler.NewServer(cfg.RunAddress, handlers.InitRoutes())

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска HTTP-сервера: %s", err.Error())
		}
	}()

	log.Infof("Сервер запущен на адресе %s", cfg.RunAddress)

	ctx, cancel := context.WithCancel(context.Background())
	go services.Orders.ProcessOrdersBackground(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Ошибка при завершении работы сервера: %s", err.Error())
	}

	cancel()

	db.Close()

	log.Info("Сервер остановлен")
}
