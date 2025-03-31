package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userID"
)

func (h *Handler) userIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "пустой заголовок авторизации")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		newErrorResponse(c, http.StatusUnauthorized, "неверный формат заголовка авторизации")
		return
	}

	token := headerParts[1]

	userID, err := h.services.Users.ParseToken(token)
	if err != nil {
		log.Errorf("Ошибка парсинга токена: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "неверный токен авторизации")
		return
	}

	c.Set(userCtx, userID)
	c.Next()
}

func getUserID(c *gin.Context) (int64, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, errors.New("пользователь не аутентифицирован")
	}

	userID, ok := id.(int64)
	if !ok {
		return 0, errors.New("неверный тип ID пользователя")
	}

	return userID, nil
}
