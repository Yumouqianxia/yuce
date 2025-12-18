package mysql

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/team"
	"backend-go/internal/core/ports"
	"gorm.io/gorm"
)

// TeamRecord 数据库模型
type TeamRecord struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Name      string `gorm:"column:name;size:100;uniqueIndex;not null"`
	ShortName string `gorm:"column:shortName;size:50"`
	LogoURL   string `gorm:"column:logoUrl;size:255"`
	IsActive  bool   `gorm:"column:isActive;not null;default:true"`
}

func (TeamRecord) TableName() string {
	return "teams"
}

// TeamRepository MySQL 实现
type TeamRepository struct {
	db *gorm.DB
}

// NewTeamRepository 创建仓储
func NewTeamRepository(db *gorm.DB) ports.TeamRepository {
	return &TeamRepository{db: db}
}

// Create 创建战队
func (r *TeamRepository) Create(ctx context.Context, t *team.Team) (*team.Team, error) {
	rec := toTeamRecord(t)
	if err := r.db.WithContext(ctx).Create(rec).Error; err != nil {
		return nil, fmt.Errorf("创建战队失败: %w", err)
	}
	return toTeamDomain(rec), nil
}

// Update 更新战队
func (r *TeamRepository) Update(ctx context.Context, id uint, t *team.Team) (*team.Team, error) {
	rec := toTeamRecord(t)
	rec.ID = id
	if err := r.db.WithContext(ctx).Model(&TeamRecord{}).Where("id = ?", id).Updates(rec).Error; err != nil {
		return nil, fmt.Errorf("更新战队失败: %w", err)
	}
	var updated TeamRecord
	if err := r.db.WithContext(ctx).First(&updated, id).Error; err != nil {
		return nil, fmt.Errorf("查询更新后的战队失败: %w", err)
	}
	return toTeamDomain(&updated), nil
}

// Delete 删除战队
func (r *TeamRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&TeamRecord{}, id).Error; err != nil {
		return fmt.Errorf("删除战队失败: %w", err)
	}
	return nil
}

// GetByID 获取战队
func (r *TeamRepository) GetByID(ctx context.Context, id uint) (*team.Team, error) {
	var rec TeamRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取战队失败: %w", err)
	}
	return toTeamDomain(&rec), nil
}

// List 列表
func (r *TeamRepository) List(ctx context.Context, includeInactive bool) ([]team.Team, error) {
	var recs []TeamRecord
	query := r.db.WithContext(ctx)
	if !includeInactive {
		query = query.Where("isActive = ?", true)
	}
	if err := query.Order("name ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("获取战队列表失败: %w", err)
	}
	result := make([]team.Team, len(recs))
	for i := range recs {
		result[i] = *toTeamDomain(&recs[i])
	}
	return result, nil
}

func toTeamRecord(t *team.Team) *TeamRecord {
	return &TeamRecord{
		ID:        t.ID,
		Name:      t.Name,
		ShortName: t.ShortName,
		LogoURL:   t.LogoURL,
		IsActive:  t.IsActive,
	}
}

func toTeamDomain(rec *TeamRecord) *team.Team {
	return &team.Team{
		ID:        rec.ID,
		Name:      rec.Name,
		ShortName: rec.ShortName,
		LogoURL:   rec.LogoURL,
		IsActive:  rec.IsActive,
	}
}
