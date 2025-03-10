package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gerfey/gophermart/internal/config"
	"github.com/Gerfey/gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour

type UserSvc struct {
	repo       repository.UserRepository
	signingKey string
	tokenTTL   time.Duration
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

func NewUserService(repo repository.UserRepository, cfg *config.Config) *UserSvc {
	return &UserSvc{
		repo:       repo,
		signingKey: cfg.JWTSigningKey,
		tokenTTL:   tokenTTL,
	}
}

func (s *UserSvc) RegisterUser(ctx context.Context, login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil && user != nil {
		return "", fmt.Errorf("пользователь с логином %s уже существует", login)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	userID, err := s.repo.CreateUser(ctx, login, string(hashedPassword))
	if err != nil {
		return "", err
	}

	token, err := s.generateToken(userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserSvc) LoginUser(ctx context.Context, login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("неверный логин или пароль")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("неверный логин или пароль")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserSvc) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи токена: %v", token.Header["alg"])
		}
		return []byte(s.signingKey), nil
	})

	if err != nil {
		return 0, fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return 0, errors.New("невалидный токен")
	}

	return claims.UserID, nil
}

func (s *UserSvc) generateToken(userID int64) (string, error) {
	claims := &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.signingKey))
}
