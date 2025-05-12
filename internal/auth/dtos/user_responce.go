package dtos

// UserResponse - данные пользователя для ответа API
type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"` // Можно форматировать время
}

// LoginResponse - ответ с токеном
type LoginResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"accessToken"`
}
