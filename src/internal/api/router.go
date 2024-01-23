package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	UserHandler *UserHandler
}

// InitRouter initialize routing information
func InitRouter(h *Handler) *gin.Engine {

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}), gin.Recovery())
	{
		r.GET("/health", getHealth)
	}
	v1 := r.Group("/v1")
	v1.Use()
	{
		v1.GET("/users/:id", h.UserHandler.GetUser)
	}

	return r
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, 1)
}
