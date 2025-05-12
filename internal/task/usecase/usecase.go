package usecase

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"task-manager/internal/task"
	"task-manager/internal/task/dtos"
	"task-manager/internal/task/entity"
	"time"
)

type taskUseCase struct {
	repo task.TaskRepository
	log  *zap.Logger
}

func NewTaskUseCase(repo task.TaskRepository, log *zap.Logger) task.TaskUseCase {
	return &taskUseCase{
		repo: repo,
		log:  log.Named("task_usecase"),
	}
}

func (uc *taskUseCase) CreateTask(ctx context.Context, req *dtos.CreateTaskRequest) (*entity.Task, error) {
	uc.log.Debug("Creating task",
		zap.String("title", req.Title),
		zap.Time("due_date", req.DueDate),
	)

	if req.Title == "" {
		uc.log.Warn("Validation failed: empty title")
		return nil, errors.New("title is required")
	}

	if req.DueDate.Before(time.Now().Add(-1 * time.Minute)) {
		uc.log.Warn("Validation failed: due date in past",
			zap.Time("due_date", req.DueDate),
		)
		return nil, errors.New("due date cannot be in the past")
	}

	task := &entity.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.StatusPending,
		Priority:    entity.Priority(req.Priority),
		DueDate:     req.DueDate,
	}

	if err := uc.repo.Create(ctx, task); err != nil {
		uc.log.Error("Failed to create task",
			zap.Error(err),
			zap.String("title", req.Title),
		)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	uc.log.Info("Task created successfully",
		zap.String("task_id", task.ID.String()),
	)
	return task, nil
}

func (uc *taskUseCase) GetTask(ctx context.Context, id string) (*entity.Task, error) {
	uc.log.Debug("Getting task", zap.String("task_id", id))

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get task",
			zap.Error(err),
			zap.String("task_id", id),
		)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	uc.log.Debug("Task retrieved",
		zap.String("task_id", id),
		zap.String("status", string(task.Status)),
	)
	return task, nil
}

func (uc *taskUseCase) UpdateTask(ctx context.Context, id string, req *dtos.UpdateTaskRequest) (*entity.Task, error) {
	uc.log.Debug("Updating task",
		zap.String("task_id", id),
		zap.Any("request", req),
	)

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Warn("Task not found for update",
			zap.String("task_id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Status != nil {
		task.Status = entity.Status(*req.Status)
	}
	if req.Priority != nil {
		task.Priority = entity.Priority(*req.Priority)
	}
	if req.DueDate != nil {
		if req.DueDate.Before(time.Now()) {
			uc.log.Warn("Invalid due date update",
				zap.Time("new_due_date", *req.DueDate),
			)
			return nil, errors.New("new due date cannot be in the past")
		}
		task.DueDate = *req.DueDate
	}

	if err := uc.repo.Update(ctx, task); err != nil {
		uc.log.Error("Failed to update task",
			zap.Error(err),
			zap.String("task_id", id),
		)
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	uc.log.Info("Task updated successfully",
		zap.String("task_id", id),
	)
	return task, nil
}

func (uc *taskUseCase) DeleteTask(ctx context.Context, id string) error {
	uc.log.Debug("Deleting task", zap.String("task_id", id))

	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.log.Error("Failed to delete task",
			zap.Error(err),
			zap.String("task_id", id),
		)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	uc.log.Info("Task deleted successfully",
		zap.String("task_id", id),
	)
	return nil
}

func (uc *taskUseCase) ListTasks(
	ctx context.Context,
	filter dtos.Filter,
	pagination dtos.Pagination,
) ([]*entity.Task, error) {
	uc.log.Debug("Listing tasks",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
	)

	if pagination.Limit <= 0 || pagination.Limit > 100 {
		pagination.Limit = 50
		uc.log.Debug("Adjusted pagination limit",
			zap.Int("new_limit", pagination.Limit),
		)
	}

	tasks, err := uc.repo.List(ctx, filter, pagination)
	if err != nil {
		uc.log.Error("Failed to list tasks",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	uc.log.Debug("Tasks listed",
		zap.Int("count", len(tasks)),
	)
	return tasks, nil
}

func (uc *taskUseCase) GetUpcomingTasks(ctx context.Context, limit int) ([]*entity.Task, error) {
	uc.log.Debug("Getting upcoming tasks",
		zap.Int("limit", limit),
	)

	if limit <= 0 || limit > 100 {
		limit = 10
		uc.log.Debug("Adjusted upcoming tasks limit",
			zap.Int("new_limit", limit),
		)
	}

	return uc.repo.List(ctx, dtos.Filter{
		Status: entity.StatusPending,
	}, dtos.Pagination{
		Limit:  limit,
		Offset: 0,
	})
}

func (uc *taskUseCase) GetOverdueTasks(ctx context.Context) ([]*entity.Task, error) {
	uc.log.Debug("Getting overdue tasks")

	threshold := time.Now().Add(-24 * time.Hour)
	tasks, err := uc.repo.GetOverdue(ctx, threshold)
	if err != nil {
		uc.log.Error("Failed to get overdue tasks",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}

	uc.log.Debug("Overdue tasks retrieved",
		zap.Int("count", len(tasks)),
	)
	return tasks, nil
}
