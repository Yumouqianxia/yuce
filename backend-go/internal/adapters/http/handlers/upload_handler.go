package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadHandler 处理文件上传
type UploadHandler struct {
	baseDir string
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(baseDir string) *UploadHandler {
	return &UploadHandler{baseDir: baseDir}
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少文件", "error": err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不支持的文件类型"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(h.baseDir, "avatars", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存文件失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename":  filename,
		"avatarUrl": "/api/uploads/avatar/" + filename,
	})
}

// GetAvatar 获取头像文件
func (h *UploadHandler) GetAvatar(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	filePath := filepath.Join(h.baseDir, "avatars", filename)
	c.File(filePath)
}
