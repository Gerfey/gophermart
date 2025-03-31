package handler

import (
	"net/http"

	"github.com/Gerfey/gophermart/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) registerUser(c *gin.Context) {
	var input model.UserCredentials

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Errorf("Ошибка разбора входящих данных при регистрации: %s", err.Error())
		newErrorResponse(c, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if input.Login == "" || input.Password == "" {
		newErrorResponse(c, http.StatusBadRequest, "логин и пароль не могут быть пустыми")
		return
	}

	token, err := h.services.Users.RegisterUser(c, input.Login, input.Password)
	if err != nil {
		log.Errorf("Ошибка регистрации пользователя: %s", err.Error())

		if err.Error() == "пользователь с логином "+input.Login+" уже существует" {
			newErrorResponse(c, http.StatusConflict, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, "ошибка при регистрации пользователя")
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *Handler) loginUser(c *gin.Context) {
	var input model.UserCredentials

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Errorf("Ошибка разбора входящих данных при входе: %s", err.Error())
		newErrorResponse(c, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if input.Login == "" || input.Password == "" {
		newErrorResponse(c, http.StatusBadRequest, "логин и пароль не могут быть пустыми")
		return
	}

	token, err := h.services.Users.LoginUser(c, input.Login, input.Password)
	if err != nil {
		log.Errorf("Ошибка аутентификации пользователя: %s", err.Error())
		newErrorResponse(c, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}
