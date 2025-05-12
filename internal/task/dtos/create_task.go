package dtos

import "time"

type CreateTaskRequest struct {
	UserID      int64
	Title       string    `json:"title" validate:"required,max=100"`
	Description *string   `json:"description,omitempty" validate:"max=500"`
	Status      string    `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress done"`
	Priority    string    `json:"priority,omitempty" validate:"omitempty,oneof=low medium high"`
	DueDate     time.Time `json:"due_date" validate:"required"`
}
