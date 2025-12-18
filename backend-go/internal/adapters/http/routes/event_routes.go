package routes

import (
	"backend-go/internal/adapters/events"
	"backend-go/internal/adapters/http/handlers"
	httpmw "backend-go/internal/adapters/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterEventRoutes 注册事件相关路由
func RegisterEventRoutes(r *gin.RouterGroup, eventManager *events.EventManager, logger *logrus.Logger, auth *httpmw.AuthMiddleware) {
	eventHandler := handlers.NewEventHandler(eventManager, logger)

	// 事件管理路由组
	events := r.Group("/events")
	{
		// 公开的统计和指标接口
		events.GET("/statistics", eventHandler.GetStatistics)
		events.GET("/metrics", eventHandler.GetMetrics)
		events.GET("/system-metrics", eventHandler.GetSystemMetrics)
		events.GET("/replay-status", eventHandler.GetReplayStatus)

		// 需要管理员权限的接口
		admin := events.Group("")
		admin.Use(auth.RequireAdmin())
		{
			admin.POST("/replay", eventHandler.ReplayEvents)
			admin.POST("/replay-failed", eventHandler.ReplayFailedEvents)
		}

		// 开发和测试接口（仅在开发环境启用）
		if gin.Mode() == gin.DebugMode {
			dev := events.Group("/dev")
			{
				dev.POST("/test", eventHandler.PublishTestEvent)
			}
		}
	}
}

// RegisterEventWebhooks 注册事件 Webhook 路由
func RegisterEventWebhooks(r *gin.RouterGroup, eventManager *events.EventManager, logger *logrus.Logger) {
	webhooks := r.Group("/webhooks/events")
	{
		// 外部系统可以通过 webhook 触发事件
		webhooks.POST("/user-action", func(c *gin.Context) {
			// 处理外部用户行为事件
			var payload map[string]interface{}
			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(400, gin.H{"error": "Invalid payload"})
				return
			}

			// 根据 payload 中的 action 类型触发相应事件
			action, ok := payload["action"].(string)
			if !ok {
				c.JSON(400, gin.H{"error": "Missing action"})
				return
			}

			userID, ok := payload["user_id"].(float64)
			if !ok {
				c.JSON(400, gin.H{"error": "Missing user_id"})
				return
			}

			switch action {
			case "page_view":
				pagePath, _ := payload["page_path"].(string)
				pageTitle, _ := payload["page_title"].(string)
				referrer, _ := payload["referrer"].(string)

				err := eventManager.PublishPageViewed(uint(userID), pagePath, pageTitle, referrer, 0)
				if err != nil {
					logger.WithError(err).Error("Failed to publish page viewed event")
					c.JSON(500, gin.H{"error": "Failed to publish event"})
					return
				}

			case "feature_use":
				featureName, _ := payload["feature_name"].(string)
				actionName, _ := payload["action_name"].(string)
				success, _ := payload["success"].(bool)

				err := eventManager.PublishFeatureUsed(uint(userID), featureName, actionName, nil, success, 0)
				if err != nil {
					logger.WithError(err).Error("Failed to publish feature used event")
					c.JSON(500, gin.H{"error": "Failed to publish event"})
					return
				}

			case "search":
				searchQuery, _ := payload["search_query"].(string)
				searchType, _ := payload["search_type"].(string)
				resultCount, _ := payload["result_count"].(float64)

				err := eventManager.PublishSearchPerformed(uint(userID), searchQuery, searchType, int(resultCount), 0)
				if err != nil {
					logger.WithError(err).Error("Failed to publish search performed event")
					c.JSON(500, gin.H{"error": "Failed to publish event"})
					return
				}

			default:
				c.JSON(400, gin.H{"error": "Unknown action"})
				return
			}

			c.JSON(200, gin.H{"message": "Event published successfully"})
		})

		// 错误报告 webhook
		webhooks.POST("/error-report", func(c *gin.Context) {
			var payload map[string]interface{}
			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(400, gin.H{"error": "Invalid payload"})
				return
			}

			userID, _ := payload["user_id"].(float64)
			errorType, _ := payload["error_type"].(string)
			errorCode, _ := payload["error_code"].(string)
			errorMessage, _ := payload["error_message"].(string)
			severity, _ := payload["severity"].(string)

			if severity == "" {
				severity = "medium"
			}

			context := make(map[string]interface{})
			if ctx, ok := payload["context"].(map[string]interface{}); ok {
				context = ctx
			}

			err := eventManager.PublishErrorEncountered(uint(userID), errorType, errorCode, errorMessage, severity, context)
			if err != nil {
				logger.WithError(err).Error("Failed to publish error encountered event")
				c.JSON(500, gin.H{"error": "Failed to publish event"})
				return
			}

			c.JSON(200, gin.H{"message": "Error event published successfully"})
		})
	}
}

// RegisterEventSSE 注册事件服务器发送事件路由
func RegisterEventSSE(r *gin.RouterGroup, eventManager *events.EventManager, logger *logrus.Logger) {
	sse := r.Group("/events/sse")
	{
		// 实时事件流
		sse.GET("/stream", func(c *gin.Context) {
			// 设置 SSE 头部
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Access-Control-Allow-Origin", "*")

			// 获取用户ID（从认证中间件）
			userID, exists := c.Get("user_id")
			if !exists {
				c.SSEvent("error", "Authentication required")
				return
			}

			// 创建事件通道
			eventChan := make(chan map[string]interface{}, 10)
			defer close(eventChan)

			// 这里应该实现实际的事件订阅逻辑
			// 为了示例，我们发送一些模拟数据
			go func() {
				// 发送连接成功消息
				eventChan <- map[string]interface{}{
					"type": "connected",
					"data": map[string]interface{}{
						"user_id": userID,
						"message": "Connected to event stream",
					},
				}

				// 定期发送心跳
				// ticker := time.NewTicker(30 * time.Second)
				// defer ticker.Stop()

				// for {
				//     select {
				//     case <-ticker.C:
				//         eventChan <- map[string]interface{}{
				//             "type": "heartbeat",
				//             "data": map[string]interface{}{
				//                 "timestamp": time.Now().Unix(),
				//             },
				//         }
				//     case <-c.Request.Context().Done():
				//         return
				//     }
				// }
			}()

			// 发送事件到客户端
			for {
				select {
				case event := <-eventChan:
					c.SSEvent(event["type"].(string), event["data"])
					c.Writer.Flush()
				case <-c.Request.Context().Done():
					logger.Info("SSE connection closed")
					return
				}
			}
		})

		// 实时统计数据流
		sse.GET("/statistics", func(c *gin.Context) {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Access-Control-Allow-Origin", "*")

			// 定期发送统计数据
			// ticker := time.NewTicker(5 * time.Second)
			// defer ticker.Stop()

			// for {
			//     select {
			//     case <-ticker.C:
			//         metrics, err := eventManager.GetSystemMetrics(c.Request.Context())
			//         if err != nil {
			//             logger.WithError(err).Error("Failed to get system metrics for SSE")
			//             continue
			//         }
			//
			//         c.SSEvent("statistics", metrics)
			//         c.Writer.Flush()
			//     case <-c.Request.Context().Done():
			//         return
			//     }
			// }

			// 为了示例，发送一次数据后关闭
			metrics, err := eventManager.GetSystemMetrics(c.Request.Context())
			if err != nil {
				c.SSEvent("error", "Failed to get metrics")
				return
			}

			c.SSEvent("statistics", metrics)
			c.Writer.Flush()
		})
	}
}
