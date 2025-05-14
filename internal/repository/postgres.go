package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnTimeout = 5 * time.Second
	defaultMaxPoolSize = 10
)

func NewPostgresDB(databaseURI string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURI)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора URI базы данных: %w", err)
	}

	cfg.MaxConns = defaultMaxPoolSize

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула соединений: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения с базой данных: %w", err)
	}

	if err := initDB(ctx, pool); err != nil {
		return nil, fmt.Errorf("ошибка инициализации схемы базы данных: %w", err)
	}

	return pool, nil
}

func initDB(ctx context.Context, pool *pgxpool.Pool) error {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	createOrdersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		number VARCHAR(255) UNIQUE NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'NEW',
		accrual FLOAT DEFAULT 0,
		uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(number)
	);`

	createBalanceTable := `
	CREATE TABLE IF NOT EXISTS balances (
		user_id INT PRIMARY KEY REFERENCES users(id),
		current FLOAT NOT NULL DEFAULT 0,
		withdrawn FLOAT NOT NULL DEFAULT 0
	);`

	createWithdrawalsTable := `
	CREATE TABLE IF NOT EXISTS withdrawals (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		order_number VARCHAR(255) NOT NULL,
		amount FLOAT NOT NULL,
		processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	if _, err := tx.Exec(ctx, createUsersTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы пользователей: %w", err)
	}

	if _, err := tx.Exec(ctx, createOrdersTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы заказов: %w", err)
	}

	if _, err := tx.Exec(ctx, createBalanceTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы баланса: %w", err)
	}

	if _, err := tx.Exec(ctx, createWithdrawalsTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы операций снятия: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}
