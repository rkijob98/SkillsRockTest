package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
	"task-manager/pkg/config"
	"task-manager/pkg/jwt"
)

func AuthMiddleware(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		log.Debug("Authorization header", zap.String("header", tokenString))

		// Извлекаем токен
		if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		} else {
			log.Error("Invalid Authorization header format")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		log.Debug("Token extracted", zap.String("token", tokenString))

		// Парсим токен
		claims, err := jwt.ParseToken(tokenString, cfg)
		if err != nil {
			log.Error("Token validation failed",
				zap.Error(err),
				zap.String("token", tokenString),
			)
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		log.Debug("Token is valid", zap.Any("claims", claims))
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
