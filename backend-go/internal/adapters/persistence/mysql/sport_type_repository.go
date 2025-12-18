package mysql

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/ports"
	"gorm.io/gorm"
)

// SportTypeRepository MySQL实现的运动类型仓储
type SportTypeRepository struct {
	db *gorm.DB
}

// NewSportTypeRepository 创建运动类型仓储实例
func NewSportTypeRepository(db *gorm.DB) *SportTypeRepository {
	return &SportTypeRepository{
		db: db,
	}
}

// Create 创建运动类型
func (r *SportTypeRepository) Create(ctx context.Context, sportType *sport.SportType) error {
	if err := r.db.WithContext(ctx).Create(sportType).Error; err != nil {
		return fmt.Errorf("failed to create sport type: %w", err)
	}
	return nil
}

// GetByID 根据ID获取运动类型
func (r *SportTypeRepository) GetByID(ctx context.Context, id uint) (*sport.SportType, error) {
	var sportType sport.SportType
	err := r.db.WithContext(ctx).
		Preload("Configuration").
		First(&sportType, id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sport type not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get sport type: %w", err)
	}
	
	return &sportType, nil
}

// GetByCode 根据代码获取运动类型
func (r *SportTypeRepository) GetByCode(ctx context.Context, code string) (*sport.SportType, error) {
	var sportType sport.SportType
	err := r.db.WithContext(ctx).
		Preload("Configuration").
		Where("code = ?", code).
		First(&sportType).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sport type not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get sport type by code: %w", err)
	}
	
	return &sportType, nil
}

// List 获取运动类型列表
func (r *SportTypeRepository) List(ctx context.Context, options *ports.ListSportTypesOptions) ([]*sport.SportType, error) {
	query := r.db.WithContext(ctx).Preload("Configuration")
	
	// 应用过滤条件
	if options.Category != "" {
		query = query.Where("category = ?", options.Category)
	}
	
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	
	// 排序
	if options.OrderBy != "" {
		query = query.Order(options.OrderBy)
	} else {
		query = query.Order("sort_order ASC, created_at ASC")
	}
	
	// 分页
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}
	
	var sportTypes []*sport.SportType
	if err := query.Find(&sportTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to list sport types: %w", err)
	}
	
	return sportTypes, nil
}

// Update 更新运动类型
func (r *SportTypeRepository) Update(ctx context.Context, sportType *sport.SportType) error {
	if err := r.db.WithContext(ctx).Save(sportType).Error; err != nil {
		return fmt.Errorf("failed to update sport type: %w", err)
	}
	return nil
}

// Delete 删除运动类型
func (r *SportTypeRepository) Delete(ctx context.Context, id uint) error {
	// 检查是否有关联的比赛
	var matchCount int64
	if err := r.db.WithContext(ctx).Table("matches").Where("sport_type_id = ?", id).Count(&matchCount).Error; err != nil {
		return fmt.Errorf("failed to check related matches: %w", err)
	}
	
	if matchCount > 0 {
		return fmt.Errorf("cannot delete sport type: %d related matches exist", matchCount)
	}
	
	// 删除运动类型（会级联删除配置）
	if err := r.db.WithContext(ctx).Delete(&sport.SportType{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete sport type: %w", err)
	}
	
	return nil
}

// Count 统计运动类型数量
func (r *SportTypeRepository) Count(ctx context.Context, options *ports.ListSportTypesOptions) (int64, error) {
	query := r.db.WithContext(ctx).Model(&sport.SportType{})
	
	// 应用过滤条件
	if options.Category != "" {
		query = query.Where("category = ?", options.Category)
	}
	
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count sport types: %w", err)
	}
	
	return count, nil
}

// CreateConfiguration 创建运动配置
func (r *SportTypeRepository) CreateConfiguration(ctx context.Context, config *sport.SportConfiguration) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create sport configuration: %w", err)
	}
	return nil
}

// UpdateConfiguration 更新运动配置
func (r *SportTypeRepository) UpdateConfiguration(ctx context.Context, config *sport.SportConfiguration) error {
	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		return fmt.Errorf("failed to update sport configuration: %w", err)
	}
	return nil
}

// GetConfiguration 获取运动配置
func (r *SportTypeRepository) GetConfiguration(ctx context.Context, sportTypeID uint) (*sport.SportConfiguration, error) {
	var config sport.SportConfiguration
	err := r.db.WithContext(ctx).
		Where("sport_type_id = ?", sportTypeID).
		First(&config).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sport configuration not found for sport type: %d", sportTypeID)
		}
		return nil, fmt.Errorf("failed to get sport configuration: %w", err)
	}
	
	return &config, nil
}

