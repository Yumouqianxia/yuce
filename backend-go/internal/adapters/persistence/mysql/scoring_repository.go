package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend-go/internal/core/domain/scoring"
	"gorm.io/gorm"
)

// MatchPointsCalculationRecord 比赛积分计算记录
type MatchPointsCalculationRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MatchID     uint      `gorm:"column:matchId;index;not null" json:"matchId"`
	Results     string    `gorm:"column:results;type:text" json:"results"` // JSON 格式存储结果
	TotalPoints int       `gorm:"column:totalPoints;not null" json:"totalPoints"`
	ProcessedAt time.Time `gorm:"column:processedAt;not null" json:"processedAt"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (MatchPointsCalculationRecord) TableName() string {
	return "match_points_calculations"
}

// PointsUpdateEventRecord 积分更新事件记录
type PointsUpdateEventRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"column:userId;index;not null" json:"userId"`
	MatchID      uint      `gorm:"column:matchId;index;not null" json:"matchId"`
	PredictionID uint      `gorm:"column:predictionId;index;not null" json:"predictionId"`
	OldPoints    int       `gorm:"column:oldPoints;not null" json:"oldPoints"`
	NewPoints    int       `gorm:"column:newPoints;not null" json:"newPoints"`
	PointsChange int       `gorm:"column:pointsChange;not null" json:"pointsChange"`
	Tournament   string    `gorm:"column:tournament;size:50;not null" json:"tournament"`
	Timestamp    time.Time `gorm:"column:timestamp;not null" json:"timestamp"`
	CreatedAt    time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (PointsUpdateEventRecord) TableName() string {
	return "points_update_events"
}

// ScoringRepository 积分计算仓储 MySQL 实现
type ScoringRepository struct {
	db *gorm.DB
}

// NewScoringRepository 创建积分计算仓储
func NewScoringRepository(db *gorm.DB) scoring.Repository {
	return &ScoringRepository{
		db: db,
	}
}

// SavePointsCalculation 保存积分计算结果
func (r *ScoringRepository) SavePointsCalculation(ctx context.Context, calculation *scoring.MatchPointsCalculation) error {
	// 序列化结果
	resultsJSON, err := json.Marshal(calculation.Results)
	if err != nil {
		return fmt.Errorf("序列化积分计算结果失败: %w", err)
	}

	record := &MatchPointsCalculationRecord{
		MatchID:     calculation.MatchID,
		Results:     string(resultsJSON),
		TotalPoints: calculation.TotalPoints,
		ProcessedAt: calculation.ProcessedAt,
	}

	// 使用 UPSERT 操作，如果记录已存在则更新
	err = r.db.WithContext(ctx).
		Where("match_id = ?", calculation.MatchID).
		Assign(record).
		FirstOrCreate(record).Error

	if err != nil {
		return fmt.Errorf("保存积分计算结果失败: %w", err)
	}

	return nil
}

// GetPointsHistory 获取积分历史
func (r *ScoringRepository) GetPointsHistory(ctx context.Context, userID uint, tournament string) ([]scoring.PointsUpdateEvent, error) {
	var records []PointsUpdateEventRecord

	query := r.db.WithContext(ctx).
		Where("userId = ?", userID).
		Order("timestamp DESC")

	if tournament != "" {
		query = query.Where("tournament = ?", tournament)
	}

	err := query.Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("获取积分历史失败: %w", err)
	}

	// 转换为领域对象
	events := make([]scoring.PointsUpdateEvent, len(records))
	for i, record := range records {
		events[i] = scoring.PointsUpdateEvent{
			UserID:       record.UserID,
			MatchID:      record.MatchID,
			PredictionID: record.PredictionID,
			OldPoints:    record.OldPoints,
			NewPoints:    record.NewPoints,
			PointsChange: record.PointsChange,
			Tournament:   record.Tournament,
			Timestamp:    record.Timestamp,
		}
	}

	return events, nil
}

// SavePointsUpdateEvent 保存积分更新事件
func (r *ScoringRepository) SavePointsUpdateEvent(ctx context.Context, event *scoring.PointsUpdateEvent) error {
	record := &PointsUpdateEventRecord{
		UserID:       event.UserID,
		MatchID:      event.MatchID,
		PredictionID: event.PredictionID,
		OldPoints:    event.OldPoints,
		NewPoints:    event.NewPoints,
		PointsChange: event.PointsChange,
		Tournament:   event.Tournament,
		Timestamp:    event.Timestamp,
	}

	err := r.db.WithContext(ctx).Create(record).Error
	if err != nil {
		return fmt.Errorf("保存积分更新事件失败: %w", err)
	}

	return nil
}

// GetMatchCalculation 获取比赛的积分计算结果
func (r *ScoringRepository) GetMatchCalculation(ctx context.Context, matchID uint) (*scoring.MatchPointsCalculation, error) {
	var record MatchPointsCalculationRecord

	err := r.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("比赛积分计算结果不存在")
		}
		return nil, fmt.Errorf("获取比赛积分计算结果失败: %w", err)
	}

	// 反序列化结果
	var results []scoring.PointsCalculationResult
	err = json.Unmarshal([]byte(record.Results), &results)
	if err != nil {
		return nil, fmt.Errorf("反序列化积分计算结果失败: %w", err)
	}

	calculation := &scoring.MatchPointsCalculation{
		MatchID:     record.MatchID,
		Results:     results,
		TotalPoints: record.TotalPoints,
		ProcessedAt: record.ProcessedAt,
	}

	return calculation, nil
}

// IsMatchProcessed 检查比赛是否已处理积分
func (r *ScoringRepository) IsMatchProcessed(ctx context.Context, matchID uint) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&MatchPointsCalculationRecord{}).
		Where("match_id = ?", matchID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("检查比赛处理状态失败: %w", err)
	}

	return count > 0, nil
}
