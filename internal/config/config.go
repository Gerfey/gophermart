package config

import (
	"flag"
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

	viper.BindEnv("RUN_ADDRESS")
	viper.BindEnv("DATABASE_URI")
	viper.BindEnv("ACCRUAL_SYSTEM_ADDRESS")
	viper.BindEnv("JWT_SIGNING_KEY")

	if cfg.RunAddress == "" {
		if address := viper.GetString("RUN_ADDRESS"); address != "" {
			cfg.RunAddress = address
		} else {
			cfg.RunAddress = ":8080"
		}
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = viper.GetString("DATABASE_URI")
	}

	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = viper.GetString("ACCRUAL_SYSTEM_ADDRESS")
	}

	if cfg.JWTSigningKey == "" {
		cfg.JWTSigningKey = viper.GetString("JWT_SIGNING_KEY")
	}

	return cfg, nil
}

func loadDotEnvFile() {
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("Не удалось загрузить .env файл из корня проекта: %v", err)
	} else {
		logrus.Info("Успешно загружен .env файл из корня проекта")
	}
}
