package mysql

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/ports"
	"gorm.io/gorm"
)

// SportScoringRuleRepository MySQL implementation of sport scoring rule repository
type SportScoringRuleRepository struct {
	db *gorm.DB
}

// NewSportScoringRuleRepository creates a new sport scoring rule repository instance
func NewSportScoringRuleRepository(db *gorm.DB) *SportScoringRuleRepository {
	return &SportScoringRuleRepository{
		db: db,
	}
}

// Create creates a scoring rule
func (r *SportScoringRuleRepository) Create(ctx context.Context, rule *sport.ScoringRule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create scoring rule: %w", err)
	}
	return nil
}

// GetByID gets a scoring rule by ID
func (r *SportScoringRuleRepository) GetByID(ctx context.Context, id uint) (*sport.ScoringRule, error) {
	var rule sport.ScoringRule
	err := r.db.WithContext(ctx).
		Preload("SportType").
		First(&rule, id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("scoring rule not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get scoring rule: %w", err)
	}
	
	return &rule, nil
}

// GetBySportTypeID gets scoring rules by sport type ID
func (r *SportScoringRuleRepository) GetBySportTypeID(ctx context.Context, sportTypeID uint) ([]*sport.ScoringRule, error) {
	var rules []*sport.ScoringRule
	err := r.db.WithContext(ctx).
		Where("sport_type_id = ?", sportTypeID).
		Order("created_at DESC").
		Find(&rules).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get scoring rules by sport type: %w", err)
	}
	
	return rules, nil
}

// GetActiveBySportTypeID gets the active scoring rule by sport type ID
func (r *SportScoringRuleRepository) GetActiveBySportTypeID(ctx context.Context, sportTypeID uint) (*sport.ScoringRule, error) {
	var rule sport.ScoringRule
	err := r.db.WithContext(ctx).
		Where("sport_type_id = ? AND is_active = ?", sportTypeID, true).
		First(&rule).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active scoring rule found for sport type: %d", sportTypeID)
		}
		return nil, fmt.Errorf("failed to get active scoring rule: %w", err)
	}
	
	return &rule, nil
}

// Update updates a scoring rule
func (r *SportScoringRuleRepository) Update(ctx context.Context, rule *sport.ScoringRule) error {
	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("failed to update scoring rule: %w", err)
	}
	return nil
}

// Delete deletes a scoring rule
func (r *SportScoringRuleRepository) Delete(ctx context.Context, id uint) error {
	// Check if it's active
	var rule sport.ScoringRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("scoring rule not found: %d", id)
		}
		return fmt.Errorf("failed to get scoring rule: %w", err)
	}
	
	if rule.IsActive {
		return fmt.Errorf("cannot delete active scoring rule")
	}
	
	// Delete the scoring rule
	if err := r.db.WithContext(ctx).Delete(&sport.ScoringRule{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete scoring rule: %w", err)
	}
	
	return nil
}

// List gets a list of scoring rules
func (r *SportScoringRuleRepository) List(ctx context.Context, options *ports.ListScoringRulesOptions) ([]*sport.ScoringRule, error) {
	query := r.db.WithContext(ctx).Preload("SportType")
	
	// Apply filters
	if options.SportTypeID != nil {
		query = query.Where("sport_type_id = ?", *options.SportTypeID)
	}
	
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	
	// Sorting
	if options.OrderBy != "" {
		query = query.Order(options.OrderBy)
	} else {
		query = query.Order("created_at DESC")
	}
	
	// Pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}
	
	var rules []*sport.ScoringRule
	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to list scoring rules: %w", err)
	}
	
	return rules, nil
}

// Count counts scoring rules
func (r *SportScoringRuleRepository) Count(ctx context.Context, options *ports.ListScoringRulesOptions) (int64, error) {
	query := r.db.WithContext(ctx).Model(&sport.ScoringRule{})
	
	// Apply filters
	if options.SportTypeID != nil {
		query = query.Where("sport_type_id = ?", *options.SportTypeID)
	}
	
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count scoring rules: %w", err)
	}
	
	return count, nil
}

// SetActive sets a scoring rule as active
func (r *SportScoringRuleRepository) SetActive(ctx context.Context, id uint) error {
	// Get the rule to activate
	var rule sport.ScoringRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("scoring rule not found: %d", id)
		}
		return fmt.Errorf("failed to get scoring rule: %w", err)
	}
	
	// Execute in transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate other rules for the same sport type
		if err := tx.Model(&sport.ScoringRule{}).
			Where("sport_type_id = ? AND id != ?", rule.SportTypeID, id).
			Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate other rules: %w", err)
		}
		
		// Activate current rule
		if err := tx.Model(&rule).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate rule: %w", err)
		}
		
		return nil
	})
}
