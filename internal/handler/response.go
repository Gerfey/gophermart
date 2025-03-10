package handler

import (
	"github.com/Gerfey/gophermart/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	log.Error(message)
	c.AbortWithStatusJSON(statusCode, model.ErrorResponse{
		Error: message,
	})
}
