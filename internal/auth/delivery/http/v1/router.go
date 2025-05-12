package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *AuthHandler) UserRoutes(router *gin.RouterGroup) {
	public := router.Group("/auth")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	//ginSwagger.WrapHandler(swaggerfiles.Handler,
	//	ginSwagger.URL("http://localhost:9000/swagger/doc.json"),
	//	ginSwagger.DefaultModelsExpandDepth(-1))
	//router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
