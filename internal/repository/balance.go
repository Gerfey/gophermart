package repository

import (
	"context"
	"fmt"

	stderrors "errors"

	"github.com/Gerfey/gophermart/internal/errors"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepo struct {
	db *pgxpool.Pool
}

func NewBalanceRepo(db *pgxpool.Pool) *BalanceRepo {
	return &BalanceRepo{db: db}
}

func (r *BalanceRepo) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	var balance model.Balance
	query := `
		SELECT user_id, current, withdrawn 
		FROM balances 
		WHERE user_id = $1
	`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
	)

	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w", errors.ErrUserBalanceNotFound)
		}
		return nil, fmt.Errorf("ошибка получения баланса пользователя: %w", err)
	}

	return &balance, nil
}

func (r *BalanceRepo) CreateBalance(ctx context.Context, userID int64) error {
	query := `
		INSERT INTO balances (user_id, current, withdrawn) 
		VALUES ($1, 0, 0) 
		ON CONFLICT (user_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("ошибка создания баланса пользователя: %w", err)
	}

	return nil
}

func (r *BalanceRepo) AddAccrual(ctx context.Context, userID int64, amount float64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	upsertQuery := `
		INSERT INTO balances (user_id, current, withdrawn) 
		VALUES ($1, $2, 0)
		ON CONFLICT (user_id) DO UPDATE 
		SET current = balances.current + $2
	`

	if _, err := tx.Exec(ctx, upsertQuery, userID, amount); err != nil {
		return fmt.Errorf("ошибка обновления баланса: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

func (r *BalanceRepo) Withdraw(ctx context.Context, userID int64, amount float64, orderNumber string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	var currentBalance float64

	balanceQuery := `
		SELECT current 
		FROM balances 
		WHERE user_id = $1
		FOR UPDATE
	`
	if err := tx.QueryRow(ctx, balanceQuery, userID).Scan(&currentBalance); err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w", errors.ErrUserBalanceNotFound)
		}
		return fmt.Errorf("ошибка получения текущего баланса: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("%w", errors.ErrInsufficientFunds)
	}

	updateBalanceQuery := `
		UPDATE balances 
		SET current = current - $1, withdrawn = withdrawn + $1 
		WHERE user_id = $2
	`
	if _, err := tx.Exec(ctx, updateBalanceQuery, amount, userID); err != nil {
		return fmt.Errorf("ошибка списания средств: %w", err)
	}

	withdrawalQuery := `
		INSERT INTO withdrawals (user_id, order_number, amount) 
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, withdrawalQuery, userID, orderNumber, amount); err != nil {
		return fmt.Errorf("ошибка создания записи о списании: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

func (r *BalanceRepo) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, amount, processed_at 
		FROM withdrawals 
		WHERE user_id = $1 
		ORDER BY processed_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории списаний: %w", err)
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var withdrawal model.Withdrawal
		if err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.OrderNumber,
			&withdrawal.Amount,
			&withdrawal.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки списания: %w", err)
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по списаниям: %w", err)
	}

	return withdrawals, nil
}
