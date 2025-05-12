package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"task-manager/internal/auth"
	"task-manager/internal/auth/dtos"
	"time"
)

type AuthHandler struct {
	uc  auth.AuthUsecase
	log *zap.Logger
}

func NewAuthHandler(uc auth.AuthUsecase, log *zap.Logger) *AuthHandler {
	return &AuthHandler{uc, log}
}

func (h *AuthHandler) Register(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()

	h.log.Info("Starting registration request",
		zap.String("path", c.FullPath()),
		zap.String("method", c.Request.Method),
	)

	var req dtos.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid registration request",
			zap.Error(err),
			zap.Any("request", req),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	user, err := h.uc.Register(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Error("Registration failed",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.Duration("duration", time.Since(startTime)),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"details": err.Error(),
		})
		return
	}

	h.log.Info("Registration successful",
		zap.Int64("user_id", user.ID),
		zap.Duration("duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()

	h.log.Info("Starting login request",
		zap.String("path", c.FullPath()),
		zap.String("method", c.Request.Method),
	)

	var req dtos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid login request",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	response, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Warn("Login failed",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.Duration("duration", time.Since(startTime)),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid credentials",
			"details": err.Error(),
		})
		return
	}

	h.log.Info("Login successful",
		zap.Int64("user_id", response.User.ID),
		zap.Duration("duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, response)
}
