package config

import (
	"cmp"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSigningKey        string
}

func LoadConfig() (*Config, error) {
	loadDotEnvFile()

	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "", "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "URI подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "адрес системы расчета начислений")
	flag.Parse()

	viper.AutomaticEnv()

	if err := viper.BindEnv("RUN_ADDRESS"); err != nil {
		return nil, fmt.Errorf("ошибка привязки переменной окружения RUN_ADDRESS: %w", err)
	}
	if err := viper.BindEnv("DATABASE_URI"); err != nil {
		return nil, fmt.Errorf("ошибка привязки переменной окружения DATABASE_URI: %w", err)
	}
	if err := viper.BindEnv("ACCRUAL_SYSTEM_ADDRESS"); err != nil {
		return nil, fmt.Errorf("ошибка привязки переменной окружения ACCRUAL_SYSTEM_ADDRESS: %w", err)
	}
	if err := viper.BindEnv("JWT_SIGNING_KEY"); err != nil {
		return nil, fmt.Errorf("ошибка привязки переменной окружения JWT_SIGNING_KEY: %w", err)
	}

	cfg.RunAddress = cmp.Or(cfg.RunAddress, viper.GetString("RUN_ADDRESS"), ":8080")
	cfg.DatabaseURI = cmp.Or(cfg.DatabaseURI, viper.GetString("DATABASE_URI"))
	cfg.AccrualSystemAddress = cmp.Or(cfg.AccrualSystemAddress, viper.GetString("ACCRUAL_SYSTEM_ADDRESS"))
	cfg.JWTSigningKey = cmp.Or(cfg.JWTSigningKey, viper.GetString("JWT_SIGNING_KEY"))

	return cfg, nil
}

func loadDotEnvFile() {
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("Не удалось загрузить .env файл из корня проекта: %v", err)
	} else {
		logrus.Info("Успешно загружен .env файл из корня проекта")
	}
}
