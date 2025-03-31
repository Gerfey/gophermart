package tests

import "errors"

var (
	ErrInvalidCredentials              = errors.New("неверные учетные данные")
	ErrOrderAlreadyUploadedByOtherUser = errors.New("заказ уже зарегистрирован другим пользователем")
	ErrInvalidOrderNumber              = errors.New("номер заказа не соответствует алгоритму Луна")
	ErrOrderNotFound                   = errors.New("заказ не найден")
)
