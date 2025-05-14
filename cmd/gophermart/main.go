package main

import (
	"context"
	"fmt"
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
	if err := run(); err != nil {
		log.Error(err)
	}
}

func run() error {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	log.Info(cfg.AccrualSystemAddress)

	db, err := repository.NewPostgresDB(cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	repos := repository.NewRepository(db)

	services := service.NewService(repos, cfg)

	handlers := handler.NewHandler(services)

	server := handler.NewServer(cfg.RunAddress, handlers.InitRoutes())

	errCh := make(chan error, 1)
	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("ошибка запуска HTTP-сервера: %w", err)
		}
	}()

	log.Infof("Сервер запущен на адресе %s", cfg.RunAddress)

	ctx, cancel := context.WithCancel(context.Background())
	go services.Orders.ProcessOrdersBackground(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	
	select {
	case <-quit:
		log.Info("Получен сигнал завершения")
	case err := <-errCh:
		cancel()
		return err
	}

	log.Info("Начинаем остановку сервера")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Ошибка при завершении работы сервера: %s", err.Error())
	}

	cancel()

	db.Close()

	log.Info("Сервер остановлен")
	return nil
}
