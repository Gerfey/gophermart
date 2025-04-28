package handler

import (
	"errors"
	"net/http"
	"strings"

	customerrors "github.com/Gerfey/gophermart/internal/errors"
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// getBalance возвращает текущий баланс пользователя.
// Метод доступен по пути GET /api/user/balance
//
// Коды ответов:
//   - 200 OK: возвращает информацию о балансе в формате JSON (текущий баланс и сумма списаний)
//   - 401 Unauthorized: пользователь не аутентифицирован
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *Handler) getBalance(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		log.Errorf("Ошибка получения ID пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "пользователь не аутентифицирован")
		return
	}

	balance, err := h.services.Balances.GetBalance(c, userID)
	if err != nil {
		log.Errorf("Ошибка получения баланса: %s", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, "ошибка получения баланса")
		return
	}

	c.JSON(http.StatusOK, balance)
}

// withdrawFromBalance обрабатывает запрос на списание средств с баланса пользователя.
// Принимает JSON с номером заказа и суммой для списания, проверяет валидность номера по алгоритму Луна
// и достаточность средств на балансе. Метод доступен по пути POST /api/user/balance/withdraw
//
// Коды ответов:
//   - 200 OK: средства успешно списаны
//   - 400 Bad Request: неверный формат запроса или некорректные данные
//   - 401 Unauthorized: пользователь не аутентифицирован
//   - 402 Payment Required: недостаточно средств на балансе
//   - 422 Unprocessable Entity: номер заказа не соответствует алгоритму Луна
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *Handler) withdrawFromBalance(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		log.Errorf("Ошибка получения ID пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "пользователь не аутентифицирован")
		return
	}

	var input model.WithdrawRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Errorf("Ошибка разбора запроса на списание: %s", err.Error())
		newErrorResponse(c, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if strings.TrimSpace(input.Order) == "" || input.Sum <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "номер заказа и сумма должны быть указаны корректно")
		return
	}

	err = h.services.Balances.Withdraw(c, userID, input.Order, input.Sum)
	if err != nil {
		log.Errorf("Ошибка списания баллов: %s", err.Error())

		if errors.Is(err, customerrors.ErrInsufficientFunds) {
			newErrorResponse(c, http.StatusPaymentRequired, err.Error())
			return
		} else if errors.Is(err, customerrors.ErrInvalidLuhn) {
			newErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, "ошибка списания баллов")
		return
	}

	c.Status(http.StatusOK)
}

// getWithdrawals возвращает историю списаний средств пользователя.
// Метод доступен по пути GET /api/user/withdrawals
//
// Коды ответов:
//   - 200 OK: возвращает список списаний в формате JSON
//   - 204 No Content: у пользователя нет списаний
//   - 401 Unauthorized: пользователь не аутентифицирован
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *Handler) getWithdrawals(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		log.Errorf("Ошибка получения ID пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "пользователь не аутентифицирован")
		return
	}

	withdrawals, err := h.services.Balances.GetWithdrawals(c, userID)
	if err != nil {
		log.Errorf("Ошибка получения истории списаний: %s", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, "ошибка получения истории списаний")
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}
