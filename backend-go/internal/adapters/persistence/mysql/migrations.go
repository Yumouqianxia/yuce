package mysql

import (
	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/user"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	// 迁移用户表
	if err := db.AutoMigrate(&user.User{}); err != nil {
		return err
	}

	// 迁移比赛表
	if err := db.AutoMigrate(&domain.Match{}); err != nil {
		return err
	}

	// 迁移预测表
	if err := db.AutoMigrate(&prediction.Prediction{}); err != nil {
		return err
	}

	// 迁移投票表
	if err := db.AutoMigrate(&prediction.Vote{}); err != nil {
		return err
	}

	// 迁移积分规则表
	if err := db.AutoMigrate(&prediction.ScoringRule{}); err != nil {
		return err
	}

	// 迁移积分计算记录表
	if err := db.AutoMigrate(&MatchPointsCalculationRecord{}); err != nil {
		return err
	}

	// 迁移积分更新事件表
	if err := db.AutoMigrate(&PointsUpdateEventRecord{}); err != nil {
		return err
	}

	// 迁移战队表
	if err := db.AutoMigrate(&TeamRecord{}); err != nil {
		return err
	}

	return nil
}
