package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/container"
	"backend-go/internal/core/services"
	"backend-go/internal/shared/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level)
	logger.Info("Starting background worker...")

	// 初始化依赖容器
	cont, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer cont.Close()

	// 初始化异步积分计算集成服务（占位，按需替换为实际依赖）
	asyncPointsIntegration := services.NewAsyncPointsIntegration(
		nil, nil, nil, nil, nil, nil, logger.GetLogger(),
	)
	defer asyncPointsIntegration.Shutdown()

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动定时任务（如需使用具体依赖，请从 cont 中获取服务/仓储并传入）
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // 每5分钟执行一次
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Background worker context cancelled")
				return
			case <-ticker.C:
				// 执行定时任务（按需接入实际逻辑）
				// executeScheduledTasks(ctx, asyncPointsIntegration, deps)
			}
		}
	}()

	// 启动积分计算状态监控
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒监控一次
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				status := asyncPointsIntegration.GetCalculationStatus()
				logger.WithFields(logrus.Fields{
					"queue_length":   status["queue_length"],
					"active_tasks":   status["active_tasks"],
					"queue_capacity": status["queue_capacity"],
				}).Debug("Async points calculation status")
			}
		}
	}()

	logger.Info("Background worker started successfully")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down background worker...")

	// 取消上下文，停止所有任务
	cancel()

	// 等待任务完成
	time.Sleep(10 * time.Second)

	logger.Info("Background worker exited")
}

// executeScheduledTasks 执行定时任务
func executeScheduledTasks(ctx context.Context, asyncPointsIntegration *services.AsyncPointsIntegration, cont *container.Container) {
	logger.Debug("Executing scheduled tasks...")

	// 1. 检查是否有已结束但未计算积分的比赛
	go checkUnprocessedMatches(ctx, asyncPointsIntegration, cont)

	// 2. 预热排行榜缓存
	go warmupLeaderboardCache(ctx, cont)

	// 3. 清理过期数据（如果需要）
	go cleanupExpiredData(ctx, cont)

	logger.Debug("Scheduled tasks initiated")
}

// checkUnprocessedMatches 检查未处理的比赛
func checkUnprocessedMatches(ctx context.Context, asyncPointsIntegration *services.AsyncPointsIntegration, cont *container.Container) {
	// 获取最近24小时内结束的比赛
	matches, err := cont.GetMatchService().GetFinishedMatches(ctx, 50)
	if err != nil {
		logger.WithError(err).Error("Failed to get finished matches")
		return
	}

	for _, match := range matches {
		// 检查是否已经计算过积分
		// 这里可以添加一个标记字段或者检查预测是否已经有积分
		predictions, err := cont.GetPredictionService().GetPredictionsByMatch(ctx, match.ID, nil)
		if err != nil {
			logger.WithError(err).WithField("match_id", match.ID).Error("Failed to get predictions for match")
			continue
		}

		// 如果有预测但没有积分，则触发计算
		needsCalculation := false
		for _, pred := range predictions {
			if pred.Prediction.EarnedPoints == 0 && !pred.Prediction.IsCorrect {
				needsCalculation = true
				break
			}
		}

		if needsCalculation {
			taskID, err := asyncPointsIntegration.ManualTriggerPointsCalculation(match.ID, nil)
			if err != nil {
				logger.WithError(err).WithField("match_id", match.ID).Error("Failed to trigger points calculation")
			} else {
				logger.WithFields(logrus.Fields{
					"match_id": match.ID,
					"task_id":  taskID,
				}).Info("Triggered points calculation for unprocessed match")
			}
		}
	}
}

// warmupLeaderboardCache 预热排行榜缓存
func warmupLeaderboardCache(ctx context.Context, cont *container.Container) {
	tournaments := []string{"SPRING", "SUMMER", "AUTUMN", "WINTER"}

	for _, tournament := range tournaments {
		if err := cont.GetLeaderboardService().RefreshLeaderboard(ctx, tournament); err != nil {
			logger.WithError(err).WithField("tournament", tournament).Warn("Failed to warmup leaderboard cache")
		} else {
			logger.WithField("tournament", tournament).Debug("Leaderboard cache warmed up")
		}
	}
}

// cleanupExpiredData 清理过期数据
func cleanupExpiredData(ctx context.Context, _ *container.Container) {
	// 这里可以实现清理逻辑
	// 例如：清理过期的缓存、日志等
	logger.Debug("Cleanup expired data task completed")
}
