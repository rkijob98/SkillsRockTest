package repository

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"task-manager/internal/auth"
	"task-manager/internal/auth/entity"
)

type authRepository struct {
	db  *sql.DB
	log *zap.Logger
}

func NewAuthRepository(db *sql.DB, log *zap.Logger) auth.AuthRepository {
	return &authRepository{db: db, log: log}
}

func (r *authRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
        INSERT INTO users (email, password)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at
    `

	r.log.Debug("Создание нового пользователя",
		zap.String("email", user.Email),
		zap.String("operation", "CreateUser"),
	)

	err := r.db.QueryRowContext(ctx, query, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.log.Error("Ошибка создания нового пользователя",
			zap.Error(err),
			zap.String("email", user.Email),
		)
		return nil, err
	}

	r.log.Info("Пользователь успешно создан",
		zap.Int64("user_id", user.ID),
	)
	return user, nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.log.Debug("Поиск пользователя по Email",
		zap.String("email", email),
		zap.String("operation", "GetUserByEmail"),
	)

	var user entity.User
	query := "SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1"

	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn("Пользователь не найден",
				zap.String("email", email),
			)
			return nil, nil
		}

		r.log.Error("Ошибка поиска пользователя",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, err
	}

	r.log.Debug("Пользователь найден",
		zap.Int64("user_id", user.ID),
	)
	return &user, nil
}
