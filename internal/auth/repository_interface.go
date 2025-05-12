package auth

import (
	"context"
	"task-manager/internal/auth/entity"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}
