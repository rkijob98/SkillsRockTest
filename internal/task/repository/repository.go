package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	redis1 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"task-manager/internal/task"
	"task-manager/internal/task/dtos"
	"task-manager/internal/task/entity"
	"task-manager/pkg/database/redis"
	"time"
)

type Repository struct {
	db    *sql.DB
	redis *redis.Client
	log   *zap.Logger
}

func NewRepository(db *sql.DB, redis *redis.Client, log *zap.Logger) task.TaskRepository {
	return &Repository{
		db:    db,
		redis: redis,
		log:   log.Named("task_repository"),
	}
}

func (r *Repository) Create(ctx context.Context, task *entity.Task) error {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	query := `
		INSERT INTO tasks (
			id, title, description, status, priority, due_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	r.log.Debug("Creating new task",
		zap.String("title", task.Title),
		zap.String("status", string(task.Status)),
	)

	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create task",
			zap.Error(err),
			zap.String("title", task.Title),
		)
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.log.Info("Task created successfully",
		zap.String("task_id", task.ID.String()),
	)
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Task, error) {
	cacheKey := fmt.Sprintf("task:%s", id)
	r.log.Debug("Fetching task from cache", zap.String("cache_key", cacheKey))

	// Попытка получить из кеша
	cachedTask, err := r.getFromCache(ctx, cacheKey)
	if err == nil && cachedTask != nil {
		r.log.Debug("Task found in cache", zap.String("task_id", id))
		return cachedTask, nil
	}

	r.log.Debug("Fetching task from database", zap.String("task_id", id))

	var task entity.Task
	query := `SELECT * FROM tasks WHERE id = $1`

	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("Task not found", zap.String("task_id", id))
			return nil, fmt.Errorf("task not found")
		}
		r.log.Error("Failed to fetch task from database",
			zap.Error(err),
			zap.String("task_id", id),
		)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	r.log.Debug("Caching task", zap.String("task_id", id))
	if err := r.cacheTask(ctx, cacheKey, &task); err != nil {
		r.log.Error("Failed to cache task",
			zap.Error(err),
			zap.String("task_id", id),
		)
	}

	return &task, nil
}

func (r *Repository) getFromCache(ctx context.Context, key string) (*entity.Task, error) {
	data, err := r.redis.Get(ctx, key)
	if err != nil {
		if err != redis1.Nil {
			r.log.Error("Cache get error",
				zap.Error(err),
				zap.String("key", key),
			)
		}
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var task entity.Task
	if err := json.Unmarshal(data, &task); err != nil {
		r.log.Error("Cache unmarshal error",
			zap.Error(err),
			zap.String("key", key),
		)
		return nil, fmt.Errorf("failed to unmarshal cached task: %w", err)
	}
	return &task, nil
}

func (r *Repository) cacheTask(ctx context.Context, key string, task *entity.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := r.redis.Set(ctx, key, data); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, task *entity.Task) error {
	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks 
		SET 
			title = $1,
			description = $2,
			status = $3,
			priority = $4,
			due_date = $5,
			updated_at = $6
		WHERE id = $7`

	r.log.Debug("Updating task",
		zap.String("task_id", task.ID.String()),
		zap.String("status", string(task.Status)),
	)

	result, err := r.db.ExecContext(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.UpdatedAt,
		task.ID,
	)

	if err != nil {
		r.log.Error("Failed to update task",
			zap.Error(err),
			zap.String("task_id", task.ID.String()),
		)
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Warn("No rows affected during update",
			zap.String("task_id", task.ID.String()),
		)
		return fmt.Errorf("task not found")
	}

	// Инвалидация кеша
	cacheKey := fmt.Sprintf("task:%s", task.ID.String())
	if err := r.redis.Delete(ctx, cacheKey); err != nil {
		r.log.Warn("Failed to invalidate cache",
			zap.Error(err),
			zap.String("cache_key", cacheKey),
		)
	}

	r.log.Info("Task updated successfully",
		zap.String("task_id", task.ID.String()),
	)
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	cacheKey := fmt.Sprintf("task:%s", id)
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	query := `DELETE FROM tasks WHERE id = $1`

	r.log.Debug("Deleting task",
		zap.String("task_id", id),
	)

	result, err := r.db.ExecContext(ctx, query, uuidID)
	if err != nil {
		r.log.Error("Failed to delete task",
			zap.Error(err),
			zap.String("task_id", id),
		)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Warn("Task not found for deletion",
			zap.String("task_id", id),
		)
		return fmt.Errorf("task not found")
	}

	// Очистка кеша
	if err := r.redis.Delete(ctx, cacheKey); err != nil {
		r.log.Warn("Failed to delete cache entry",
			zap.Error(err),
			zap.String("cache_key", cacheKey),
		)
	}

	r.log.Info("Task deleted successfully",
		zap.String("task_id", id),
	)
	return nil
}

func (r *Repository) List(
	ctx context.Context,
	filter dtos.Filter,
	pagination dtos.Pagination,
) ([]*entity.Task, error) {
	baseQuery := "SELECT * FROM tasks WHERE 1=1"
	args := []interface{}{}
	argCounter := 1

	// Формирование условий фильтра
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argCounter)
		args = append(args, filter.Status)
		argCounter++
	}
	if filter.Priority != "" {
		baseQuery += fmt.Sprintf(" AND priority = $%d", argCounter)
		args = append(args, filter.Priority)
		argCounter++
	}
	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND title ILIKE $%d", argCounter)
		args = append(args, "%"+filter.Search+"%")
		argCounter++
	}

	// Добавляем сортировку и пагинацию
	baseQuery += " ORDER BY due_date ASC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, pagination.Limit, pagination.Offset)

	r.log.Debug("Listing tasks",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
	)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		r.log.Error("Failed to list tasks",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan task row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error during rows iteration",
			zap.Error(err),
		)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	r.log.Debug("Tasks listed successfully",
		zap.Int("count", len(tasks)),
	)
	return tasks, nil
}

func (r *Repository) GetOverdue(ctx context.Context, threshold time.Time) ([]*entity.Task, error) {
	query := `
		SELECT * 
		FROM tasks 
		WHERE due_date < $1 
		AND status != $2 
		ORDER BY due_date ASC`

	r.log.Debug("Fetching overdue tasks",
		zap.Time("threshold", threshold),
	)

	rows, err := r.db.QueryContext(ctx, query, threshold, entity.StatusDone)
	if err != nil {
		r.log.Error("Failed to fetch overdue tasks",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan overdue task",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error during overdue tasks iteration",
			zap.Error(err),
		)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	r.log.Debug("Overdue tasks fetched",
		zap.Int("count", len(tasks)),
	)
	return tasks, nil
}
