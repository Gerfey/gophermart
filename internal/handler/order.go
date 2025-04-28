package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// createOrder обрабатывает запрос на создание нового заказа.
// Принимает номер заказа в теле запроса в текстовом формате, проверяет его валидность по алгоритму Луна
// и регистрирует в системе. Метод доступен по пути POST /api/user/orders
//
// Коды ответов:
//   - 200 OK: заказ уже был зарегистрирован ранее этим же пользователем
//   - 202 Accepted: новый заказ принят в обработку
//   - 400 Bad Request: неверный формат запроса или пустой номер заказа
//   - 401 Unauthorized: пользователь не аутентифицирован
//   - 409 Conflict: заказ уже зарегистрирован другим пользователем
//   - 422 Unprocessable Entity: номер заказа не соответствует алгоритму Луна
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *Handler) createOrder(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		log.Errorf("Ошибка получения ID пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "пользователь не аутентифицирован")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("Ошибка чтения тела запроса: %s", err.Error())
		newErrorResponse(c, http.StatusBadRequest, "ошибка чтения тела запроса")
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		newErrorResponse(c, http.StatusBadRequest, "номер заказа не может быть пустым")
		return
	}

	code, err := h.services.Orders.CreateOrder(c, userID, orderNumber)
	if err != nil {
		log.Errorf("Ошибка создания заказа: %s", err.Error())
		newErrorResponse(c, code, err.Error())
		return
	}

	c.Status(code)
}

// getOrders возвращает список заказов текущего пользователя.
// Метод доступен по пути GET /api/user/orders
//
// Коды ответов:
//   - 200 OK: возвращает список заказов в формате JSON
//   - 204 No Content: у пользователя нет заказов
//   - 401 Unauthorized: пользователь не аутентифицирован
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *Handler) getOrders(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		log.Errorf("Ошибка получения ID пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "пользователь не аутентифицирован")
		return
	}

	orders, err := h.services.Orders.GetOrdersByUserID(c, userID)
	if err != nil {
		log.Errorf("Ошибка получения списка заказов: %s", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, "ошибка получения списка заказов")
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
