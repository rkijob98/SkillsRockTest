package dtos

import "time"

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,max=100"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=500"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress done"`
	Priority    *string    `json:"priority,omitempty" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}
