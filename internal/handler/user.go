package handler

import (
	"net/http"

	"github.com/Gerfey/gophermart/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// registerUser обрабатывает запрос на регистрацию нового пользователя.
// Принимает JSON с логином и паролем, создает нового пользователя и возвращает JWT токен.
// Метод доступен по пути POST /api/user/register
//
// Коды ответов:
//   - 200 OK: пользователь успешно зарегистрирован, в заголовке Authorization возвращается токен
//   - 400 Bad Request: неверный формат запроса или пустые логин/пароль
//   - 409 Conflict: пользователь с таким логином уже существует
//   - 500 Internal Server Error: внутренняя ошибка сервера
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

// loginUser обрабатывает запрос на аутентификацию пользователя.
// Принимает JSON с логином и паролем, проверяет учетные данные и возвращает JWT токен.
// Метод доступен по пути POST /api/user/login
//
// Коды ответов:
//   - 200 OK: аутентификация успешна, в заголовке Authorization возвращается токен
//   - 400 Bad Request: неверный формат запроса или пустые логин/пароль
//   - 401 Unauthorized: неверный логин или пароль
//   - 500 Internal Server Error: внутренняя ошибка сервера
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
