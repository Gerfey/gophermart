package errors

import (
	"errors"
)

var (
	ErrInsufficientFunds   = errors.New("недостаточно средств")
	ErrUserBalanceNotFound = errors.New("баланс пользователя не найден")
	ErrInvalidLuhn         = errors.New("номер заказа не соответствует алгоритму Луна")
	ErrOrderAlreadyExists  = errors.New("заказ уже зарегистрирован другим пользователем")
)
