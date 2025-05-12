package dtos

import "task-manager/internal/task/entity"

type Filter struct {
	Status   entity.Status
	Priority entity.Priority
	Search   string
}
