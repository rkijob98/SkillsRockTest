package v1

import (
	"github.com/gin-gonic/gin"
)

func (h *TaskHandler) TaskRoutes(router *gin.RouterGroup, auth gin.HandlerFunc) {
	// Группа защиущенных роутов для задач
	taskGroup := router.Group("/tasks").Use(auth)
	{
		taskGroup.POST("", h.CreateTask)
		taskGroup.GET("/:id", h.GetTask)
		taskGroup.PUT("/:id", h.UpdateTask)
		taskGroup.DELETE("/:id", h.DeleteTask)
		taskGroup.GET("", h.ListTasks)
	}
}
