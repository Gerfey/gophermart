package service

import (
	"context"
	"fmt"

	"github.com/Gerfey/gophermart/internal/errors"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/Gerfey/gophermart/internal/repository"
)

type BalanceSvc struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceSvc {
	return &BalanceSvc{
		repo: repo,
	}
}

func (s *BalanceSvc) GetBalance(ctx context.Context, userID int64) (model.BalanceResponse, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return model.BalanceResponse{}, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	response := model.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	return response, nil
}

func (s *BalanceSvc) Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	if !IsValidLuhnNumber(orderNumber) {
		return fmt.Errorf("%w", errors.ErrInvalidLuhn)
	}

	err := s.repo.Withdraw(ctx, userID, amount, orderNumber)
	if err != nil {
		return fmt.Errorf("ошибка списания баллов: %w", err)
	}

	return nil
}

func (s *BalanceSvc) GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalResponse, error) {
	withdrawals, err := s.repo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории списаний: %w", err)
	}

	response := make([]model.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		response = append(response, model.WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         w.Amount,
			ProcessedAt: w.ProcessedAt,
		})
	}

	return response, nil
}
