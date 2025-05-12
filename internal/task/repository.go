package task

import (
	"context"
	"task-manager/internal/task/dtos"
	"task-manager/internal/task/entity"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id string) (*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id string) error
	List(
		ctx context.Context,
		filter dtos.Filter,
		pagination dtos.Pagination,
	) ([]*entity.Task, error)
	GetOverdue(ctx context.Context, threshold time.Time) ([]*entity.Task, error)
}
