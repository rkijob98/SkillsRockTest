package usecase

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"task-manager/internal/auth"
	"task-manager/internal/auth/dtos"
	"task-manager/internal/auth/entity"
	"task-manager/pkg/config"
	"task-manager/pkg/jwt"
	"time"
)

type authUsecase struct {
	repo auth.AuthRepository
	cfg  *config.Config
	log  *zap.Logger
}

func NewAuthUsecase(repo auth.AuthRepository, cfg *config.Config, log *zap.Logger) auth.AuthUsecase {
	return &authUsecase{repo, cfg, log}
}

func (uc *authUsecase) Register(ctx context.Context, email, password string) (*dtos.UserResponse, error) {
	uc.log.Info("Starting user registration",
		zap.String("email", email),
		zap.String("operation", "Register"),
	)

	startTime := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("Failed to hash password",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	createdUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		uc.log.Error("Failed to create user in repository",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	uc.log.Info("User successfully registered",
		zap.Int64("user_id", createdUser.ID),
		zap.String("email", createdUser.Email),
		zap.Duration("duration", time.Since(startTime)),
	)

	return &dtos.UserResponse{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *authUsecase) Login(ctx context.Context, email, password string) (*dtos.LoginResponse, error) {
	uc.log.Info("Starting user login",
		zap.String("email", email),
		zap.String("operation", "Login"),
	)

	startTime := time.Now()
	user, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		uc.log.Error("Failed to get user from repository",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	if user == nil {
		uc.log.Warn("User not found",
			zap.String("email", email),
		)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		uc.log.Warn("Invalid password attempt",
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, errors.New("invalid credentials")
	}

	token, err := jwt.GenerateToken(user.ID, uc.cfg)
	if err != nil {
		uc.log.Error("Failed to generate JWT token",
			zap.Error(err),
			zap.Int64("user_id", user.ID),
		)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	uc.log.Info("User successfully logged in",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("accessToken", token),
		zap.Duration("duration", time.Since(startTime)),
	)

	return &dtos.LoginResponse{
		User: dtos.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
		AccessToken: token,
	}, nil
}
