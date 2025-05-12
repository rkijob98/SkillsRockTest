package task

import (
	"context"
	"task-manager/internal/task/dtos"
	"task-manager/internal/task/entity"
)

type TaskUseCase interface {
	CreateTask(ctx context.Context, req *dtos.CreateTaskRequest) (*entity.Task, error)
	GetTask(ctx context.Context, id string) (*entity.Task, error)
	UpdateTask(ctx context.Context, id string, req *dtos.UpdateTaskRequest) (*entity.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filter dtos.Filter, pagination dtos.Pagination) ([]*entity.Task, error)
	GetUpcomingTasks(ctx context.Context, limit int) ([]*entity.Task, error)
	GetOverdueTasks(ctx context.Context) ([]*entity.Task, error)
}
