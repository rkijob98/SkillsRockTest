package auth

import (
	"context"
	"task-manager/internal/auth/dtos"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password string) (*dtos.UserResponse, error)
	Login(ctx context.Context, email, password string) (*dtos.LoginResponse, error)
}
