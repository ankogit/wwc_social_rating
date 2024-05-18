package handler

import (
	"github.com/ankogit/wwc_social_rating/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		api.GET("/test", h.getTestPage)
		api.GET("/stats/image/:id/file.png", h.getTestPage)
	}

	return router
}
