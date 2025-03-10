package model

import (
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type AccrualSystemStatus string

const (
	AccrualStatusRegistered AccrualSystemStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualSystemStatus = "INVALID"
	AccrualStatusProcessing AccrualSystemStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualSystemStatus = "PROCESSED"
)

type Order struct {
	ID         int64       `db:"id"`
	UserID     int64       `db:"user_id"`
	Number     string      `db:"number"`
	Status     OrderStatus `db:"status"`
	Accrual    float64     `db:"accrual"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    float64     `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

type Withdrawal struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Amount      float64   `db:"amount"`
	ProcessedAt time.Time `db:"processed_at"`
}

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Balance struct {
	UserID    int64   `db:"user_id"`
	Current   float64 `db:"current"`
	Withdrawn float64 `db:"withdrawn"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required,gt=0"`
}

type UserCredentials struct {
	Login    string `json:"login" binding:"required,min=1"`
	Password string `json:"password" binding:"required,min=1"`
}

type AccrualResponse struct {
	Order   string              `json:"order"`
	Status  AccrualSystemStatus `json:"status"`
	Accrual float64             `json:"accrual,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
