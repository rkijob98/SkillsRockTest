package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"task-manager/internal/task"
	"task-manager/internal/task/dtos"
	"task-manager/internal/task/entity"
)

type TaskHandler struct {
	uc  task.TaskUseCase
	log *zap.Logger
}

func NewTaskHandler(uc task.TaskUseCase, log *zap.Logger) *TaskHandler {
	return &TaskHandler{
		uc:  uc,
		log: log.Named("task_handler"),
	}
}

// CreateTask создает новую задачу
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dtos.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Получаем user_id как int64
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Преобразуем к нужному типу
	uid, ok := userID.(int64)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Передаем в usecase (если нужно)
	req.UserID = uid

	task, err := h.uc.CreateTask(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask возвращает задачу по ID
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := h.uc.GetTask(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Task not found", zap.Error(err), zap.String("task_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask обновляет существующую задачу
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req dtos.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	task, err := h.uc.UpdateTask(c.Request.Context(), id, &req)
	if err != nil {
		h.log.Error("Failed to update task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask удаляет задачу
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.DeleteTask(c.Request.Context(), id); err != nil {
		h.log.Error("Failed to delete task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Deletion failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTasks возвращает список задач с фильтрацией
func (h *TaskHandler) ListTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tasks, err := h.uc.ListTasks(c.Request.Context(), dtos.Filter{
		Status: entity.Status(c.Query("status")),
	}, dtos.Pagination{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		h.log.Error("Failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
