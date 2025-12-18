package routes

import (
	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/prediction"
	"github.com/gin-gonic/gin"
)

// PredictionRoutes 预测路由
type PredictionRoutes struct {
	predictionHandler *handlers.PredictionHandler
	authMiddleware    *middleware.AuthMiddleware
}

// NewPredictionRoutes 创建预测路由
func NewPredictionRoutes(predictionService prediction.Service, authMiddleware *middleware.AuthMiddleware) *PredictionRoutes {
	return &PredictionRoutes{
		predictionHandler: handlers.NewPredictionHandler(predictionService),
		authMiddleware:    authMiddleware,
	}
}

// RegisterRoutes 注册预测路由
func (r *PredictionRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	predictions := rg.Group("/predictions")

	// 公开路由 - 不需要认证
	{
		predictions.GET("", r.predictionHandler.GetPredictionsByMatch)           // 获取比赛预测列表
		predictions.GET("/:id", r.predictionHandler.GetPrediction)               // 获取预测详情
		predictions.GET("/featured", r.predictionHandler.GetFeaturedPredictions) // 获取精选预测
	}

	// 需要认证的路由
	authenticated := predictions.Group("")
	authenticated.Use(r.authMiddleware.RequireAuth())
	{
		// 预测管理
		authenticated.POST("", r.predictionHandler.CreatePrediction)     // 创建预测
		authenticated.PUT("/:id", r.predictionHandler.UpdatePrediction)  // 更新预测
		authenticated.GET("/my", r.predictionHandler.GetUserPredictions) // 获取用户预测列表
		authenticated.GET("/my-predictions", r.predictionHandler.GetUserPredictions) // 兼容旧路径
		authenticated.POST("/reverify/:id", r.predictionHandler.ReverifyPrediction) // 重新验证预测

		// 投票功能
		authenticated.POST("/:id/vote", r.predictionHandler.VotePrediction)     // 投票支持预测
		authenticated.DELETE("/:id/vote", r.predictionHandler.UnvotePrediction) // 取消投票
	}
}
