package handler

import (
	"context"
	"net/http"
	"time"

	"ai-ops/internal/auth"
	"ai-ops/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

// AuthHandler 认证处理器
type AuthHandler struct {
	jwtManager *auth.JWTManager
	userRepo   UserRepository
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
	}
}

// NewAuthHandlerWithRepo 创建带用户仓库的认证处理器
func NewAuthHandlerWithRepo(jwtManager *auth.JWTManager, userRepo UserRepository) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		userRepo:   userRepo,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
}

// RefreshTokenRequest 刷新 token 请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证用户凭据
	user, err := h.validateCredentials(c.Request.Context(), req.Username, req.Password)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户已被禁用",
		})
		return
	}

	// 生成 JWT token
	token, err := h.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成令牌失败",
		})
		return
	}

	// 更新最后登录时间
	if h.userRepo != nil {
		_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)
	}

	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			UserID:    user.ID,
			Username:  user.Username,
		},
	})
}

// RefreshToken 刷新 token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	newToken, err := h.jwtManager.Refresh(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "令牌刷新失败: " + err.Error(),
		})
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "令牌刷新成功",
		"data": LoginResponse{
			Token:     newToken,
			ExpiresAt: expiresAt,
		},
	})
}

// ValidateToken 验证 token
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未提供认证令牌",
		})
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := h.jwtManager.Verify(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "令牌无效: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "令牌有效",
		"data": gin.H{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"expires":  claims.ExpiresAt.Unix(),
		},
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// validateCredentials 验证用户凭据
func (h *AuthHandler) validateCredentials(ctx context.Context, username, password string) (*model.User, error) {
	// 优先使用数据库验证
	if h.userRepo != nil {
		user, err := h.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, err
		}
		return user, nil
	}

	// 回退到演示账户（仅开发环境）
	demoUsers := map[string]string{
		"admin":    "admin123",
		"operator": "operator123",
	}

	storedPassword, exists := demoUsers[username]
	if !exists || password != storedPassword {
		return nil, nil
	}

	return &model.User{
		ID:       username,
		Username: username,
		Role:     "admin",
		Status:   "active",
	}, nil
}

// HashPassword 生成密码哈希（工具函数）
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
