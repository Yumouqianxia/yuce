// Package mysql provides MySQL-specific implementations of repository interfaces.
//
// This file implements the migration repository for managing database schema
// migrations using GORM and MySQL.
package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/core/domain"
	"backend-go/pkg/database"
)

// MigrationRepository implements the migration repository interface for MySQL.
type MigrationRepository struct {
	db *database.DB
}

// NewMigrationRepository creates a new migration repository instance.
func NewMigrationRepository(db *database.DB) *MigrationRepository {
	return &MigrationRepository{
		db: db,
	}
}

// CreateMigrationsTable creates the migrations table if it doesn't exist.
func (r *MigrationRepository) CreateMigrationsTable(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&domain.Migration{})
}

// CreateSeedDataTable creates the seed data table if it doesn't exist.
func (r *MigrationRepository) CreateSeedDataTable(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&domain.SeedData{})
}

// GetAppliedMigrations returns all successfully applied migrations ordered by version.
func (r *MigrationRepository) GetAppliedMigrations(ctx context.Context) ([]domain.Migration, error) {
	var migrations []domain.Migration
	
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.MigrationStatusCompleted).
		Order("version ASC").
		Find(&migrations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	return migrations, nil
}

// GetMigrationByVersion returns a migration by its version.
func (r *MigrationRepository) GetMigrationByVersion(ctx context.Context, version string) (*domain.Migration, error) {
	var migration domain.Migration
	
	err := r.db.WithContext(ctx).
		Where("version = ?", version).
		First(&migration).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get migration by version %s: %w", version, err)
	}
	
	return &migration, nil
}

// SaveMigration saves a migration record to the database.
func (r *MigrationRepository) SaveMigration(ctx context.Context, migration *domain.Migration) error {
	err := r.db.WithContext(ctx).Save(migration).Error
	if err != nil {
		return fmt.Errorf("failed to save migration %s: %w", migration.Version, err)
	}
	
	return nil
}

// CreateMigration creates a new migration record.
func (r *MigrationRepository) CreateMigration(ctx context.Context, migration *domain.Migration) error {
	err := r.db.WithContext(ctx).Create(migration).Error
	if err != nil {
		return fmt.Errorf("failed to create migration %s: %w", migration.Version, err)
	}
	
	return nil
}

// UpdateMigrationStatus updates the status of a migration.
func (r *MigrationRepository) UpdateMigrationStatus(ctx context.Context, version string, status domain.MigrationStatus) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Migration{}).
		Where("version = ?", version).
		Update("status", status).Error
	
	if err != nil {
		return fmt.Errorf("failed to update migration status for version %s: %w", version, err)
	}
	
	return nil
}

// GetPendingMigrations returns all pending migrations ordered by version.
func (r *MigrationRepository) GetPendingMigrations(ctx context.Context) ([]domain.Migration, error) {
	var migrations []domain.Migration
	
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.MigrationStatusPending).
		Order("version ASC").
		Find(&migrations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}
	
	return migrations, nil
}

// GetFailedMigrations returns all failed migrations.
func (r *MigrationRepository) GetFailedMigrations(ctx context.Context) ([]domain.Migration, error) {
	var migrations []domain.Migration
	
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.MigrationStatusFailed).
		Order("version DESC").
		Find(&migrations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get failed migrations: %w", err)
	}
	
	return migrations, nil
}

// GetLastAppliedMigration returns the most recently applied migration.
func (r *MigrationRepository) GetLastAppliedMigration(ctx context.Context) (*domain.Migration, error) {
	var migration domain.Migration
	
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.MigrationStatusCompleted).
		Order("version DESC").
		First(&migration).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last applied migration: %w", err)
	}
	
	return &migration, nil
}

// DeleteMigration deletes a migration record.
func (r *MigrationRepository) DeleteMigration(ctx context.Context, version string) error {
	err := r.db.WithContext(ctx).
		Where("version = ?", version).
		Delete(&domain.Migration{}).Error
	
	if err != nil {
		return fmt.Errorf("failed to delete migration %s: %w", version, err)
	}
	
	return nil
}

// GetMigrationHistory returns the complete migration history.
func (r *MigrationRepository) GetMigrationHistory(ctx context.Context, limit int) ([]domain.Migration, error) {
	var migrations []domain.Migration
	
	query := r.db.WithContext(ctx).Order("executed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&migrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}
	
	return migrations, nil
}

// GetSeedDataByName returns seed data by name.
func (r *MigrationRepository) GetSeedDataByName(ctx context.Context, name string) (*domain.SeedData, error) {
	var seedData domain.SeedData
	
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&seedData).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get seed data by name %s: %w", name, err)
	}
	
	return &seedData, nil
}

// SaveSeedData saves seed data record to the database.
func (r *MigrationRepository) SaveSeedData(ctx context.Context, seedData *domain.SeedData) error {
	err := r.db.WithContext(ctx).Save(seedData).Error
	if err != nil {
		return fmt.Errorf("failed to save seed data %s: %w", seedData.Name, err)
	}
	
	return nil
}

// GetAppliedSeedData returns all applied seed data.
func (r *MigrationRepository) GetAppliedSeedData(ctx context.Context) ([]domain.SeedData, error) {
	var seedData []domain.SeedData
	
	err := r.db.WithContext(ctx).
		Where("applied = ?", true).
		Order("applied_at ASC").
		Find(&seedData).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get applied seed data: %w", err)
	}
	
	return seedData, nil
}

// ExecuteInTransaction executes a function within a database transaction.
func (r *MigrationRepository) ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// ExecuteRawSQL executes raw SQL statements.
func (r *MigrationRepository) ExecuteRawSQL(ctx context.Context, sql string) error {
	err := r.db.WithContext(ctx).Exec(sql).Error
	if err != nil {
		return fmt.Errorf("failed to execute raw SQL: %w", err)
	}
	
	return nil
}

// GetDatabaseVersion returns the current database version.
func (r *MigrationRepository) GetDatabaseVersion(ctx context.Context) (string, error) {
	lastMigration, err := r.GetLastAppliedMigration(ctx)
	if err != nil {
		return "", err
	}
	
	if lastMigration == nil {
		return "0", nil
	}
	
	return lastMigration.Version, nil
}

// CheckMigrationLock checks if there's an active migration lock.
func (r *MigrationRepository) CheckMigrationLock(ctx context.Context) (bool, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&domain.Migration{}).
		Where("status = ?", domain.MigrationStatusRunning).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("failed to check migration lock: %w", err)
	}
	
	return count > 0, nil
}

// ClearStaleLocks clears migration locks that are older than the specified duration.
func (r *MigrationRepository) ClearStaleLocks(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	
	err := r.db.WithContext(ctx).
		Model(&domain.Migration{}).
		Where("status = ? AND executed_at < ?", domain.MigrationStatusRunning, cutoff).
		Update("status", domain.MigrationStatusFailed).Error
	
	if err != nil {
		return fmt.Errorf("failed to clear stale locks: %w", err)
	}
	
	return nil
}