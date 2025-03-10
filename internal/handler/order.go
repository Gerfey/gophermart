package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

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
