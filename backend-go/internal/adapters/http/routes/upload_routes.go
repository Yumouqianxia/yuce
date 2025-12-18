package routes

import (
	"os"
	"path/filepath"

	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"github.com/gin-gonic/gin"
)

// UploadRoutes 上传路由
type UploadRoutes struct {
	handler        *handlers.UploadHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewUploadRoutes 创建上传路由
func NewUploadRoutes(auth *middleware.AuthMiddleware, baseDir string) *UploadRoutes {
	// 确保目录存在
	avatarDir := filepath.Join(baseDir, "avatars")
	_ = os.MkdirAll(avatarDir, os.ModePerm)

	return &UploadRoutes{
		handler:        handlers.NewUploadHandler(baseDir),
		authMiddleware: auth,
	}
}

// RegisterRoutes 注册上传路由
func (r *UploadRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	uploads := rg.Group("/uploads")

	// 公开获取头像
	uploads.GET("/avatar/:filename", r.handler.GetAvatar)

	// 上传需要认证
	uploadsAuth := uploads.Group("")
	uploadsAuth.Use(r.authMiddleware.RequireAuth())
	{
		uploadsAuth.POST("/avatar", r.handler.UploadAvatar)
	}
}
