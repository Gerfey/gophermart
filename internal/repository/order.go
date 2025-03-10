package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gerfey/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) CreateOrder(ctx context.Context, userID int64, number string) (int64, error) {
	var id int64
	query := `
		INSERT INTO orders (user_id, number, status) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, userID, number, model.OrderStatusNew).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания заказа: %w", err)
	}

	return id, nil
}

func (r *OrderRepo) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	var order model.Order
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at 
		FROM orders 
		WHERE number = $1
	`

	err := r.db.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения заказа: %w", err)
	}

	return &order, nil
}

func (r *OrderRepo) GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов пользователя: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки заказа: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по заказам: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus) error {
	query := `
		UPDATE orders 
		SET status = $1 
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заказа: %w", err)
	}

	return nil
}

func (r *OrderRepo) UpdateOrderAccrual(ctx context.Context, orderID int64, accrual float64) error {
	query := `
		UPDATE orders 
		SET accrual = $1, status = $2 
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, accrual, model.OrderStatusProcessed, orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления начисления заказа: %w", err)
	}

	return nil
}

func (r *OrderRepo) GetNewOrProcessingOrders(ctx context.Context) ([]*model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at 
		FROM orders 
		WHERE status = $1 OR status = $2
	`

	rows, err := r.db.Query(ctx, query, model.OrderStatusNew, model.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов для проверки: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки заказа: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по заказам: %w", err)
	}

	return orders, nil
}
