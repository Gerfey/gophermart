package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gerfey/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	query := `
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, login, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	balanceQuery := `
		INSERT INTO balances (user_id, current, withdrawn) 
		VALUES ($1, 0, 0)
	`
	_, err = r.db.Exec(ctx, balanceQuery, id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания баланса пользователя: %w", err)
	}

	return id, nil
}

func (r *UserRepo) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, login, password_hash, created_at 
		FROM users 
		WHERE login = $1
	`

	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, login, password_hash, created_at 
		FROM users 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return &user, nil
}
