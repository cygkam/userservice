package api

import (
	"fmt"
	"net/http"

	"user-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	UserService *service.UserService
}

// @Router /v1/users/:id [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	_, err := uuid.Parse(id)
	if err != nil {
		sendAbortWithStatusJSON(c, http.StatusBadRequest, fmt.Sprintf("Couldn't parse uuid %v", err.Error()))
		return
	}

	user, err := h.UserService.GetUser(c, id)
	if err != nil {
		sendAbortWithStatusJSON(c, http.StatusBadRequest, fmt.Sprintf("Error occured: %v", err.Error()))
		return
	}

	c.JSON(200, user)
}

func sendAbortWithStatusJSON(c *gin.Context, httpCode int, message string) {
	c.AbortWithStatusJSON(httpCode,
		gin.H{
			"http-code": httpCode,
			"cause":     message,
		})
	logrus.Error(message)
}
