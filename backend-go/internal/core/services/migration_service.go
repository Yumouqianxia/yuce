// Package services contains the core business logic services.
//
// This file implements the migration service for managing database schema
// migrations with support for both GORM AutoMigrate and manual SQL migrations.
package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/sirupsen/logrus"

	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/database"
)

// MigrationRepository defines the interface for migration data access.
type MigrationRepository interface {
	CreateMigrationsTable(ctx context.Context) error
	CreateSeedDataTable(ctx context.Context) error
	GetAppliedMigrations(ctx context.Context) ([]domain.Migration, error)
	GetMigrationByVersion(ctx context.Context, version string) (*domain.Migration, error)
	SaveMigration(ctx context.Context, migration *domain.Migration) error
	CreateMigration(ctx context.Context, migration *domain.Migration) error
	UpdateMigrationStatus(ctx context.Context, version string, status domain.MigrationStatus) error
	GetPendingMigrations(ctx context.Context) ([]domain.Migration, error)
	GetFailedMigrations(ctx context.Context) ([]domain.Migration, error)
	GetLastAppliedMigration(ctx context.Context) (*domain.Migration, error)
	DeleteMigration(ctx context.Context, version string) error
	GetMigrationHistory(ctx context.Context, limit int) ([]domain.Migration, error)
	GetSeedDataByName(ctx context.Context, name string) (*domain.SeedData, error)
	SaveSeedData(ctx context.Context, seedData *domain.SeedData) error
	GetAppliedSeedData(ctx context.Context) ([]domain.SeedData, error)
	ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error
	ExecuteRawSQL(ctx context.Context, sql string) error
	GetDatabaseVersion(ctx context.Context) (string, error)
	CheckMigrationLock(ctx context.Context) (bool, error)
	ClearStaleLocks(ctx context.Context, maxAge time.Duration) error
}

// MigrationService handles database migration operations.
type MigrationService struct {
	db         *database.DB
	repository MigrationRepository
	logger     *logrus.Logger
}

// NewMigrationService creates a new migration service instance.
func NewMigrationService(db *database.DB, repository MigrationRepository) *MigrationService {
	return &MigrationService{
		db:         db,
		repository: repository,
		logger:     logger.GetLogger(),
	}
}

// MigrationFile represents a migration file.
type MigrationFile struct {
	Version   string
	Name      string
	Type      domain.MigrationType
	FilePath  string
	Content   string
	Checksum  string
}

// InitializeMigrationSystem initializes the migration system.
func (s *MigrationService) InitializeMigrationSystem(ctx context.Context) error {
	s.logger.Info("Initializing migration system...")

	// Create migration tables
	if err := s.repository.CreateMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if err := s.repository.CreateSeedDataTable(ctx); err != nil {
		return fmt.Errorf("failed to create seed data table: %w", err)
	}

	// Clear stale locks (older than 1 hour)
	if err := s.repository.ClearStaleLocks(ctx, time.Hour); err != nil {
		s.logger.Warn("Failed to clear stale migration locks: %v", err)
	}

	s.logger.Info("Migration system initialized successfully")
	return nil
}

// AutoMigrate performs GORM auto-migration for all models.
func (s *MigrationService) AutoMigrate(ctx context.Context) error {
	s.logger.Info("Starting GORM auto-migration...")

	// Define all models that need to be migrated
	models := []interface{}{
		&user.User{},
		&domain.Match{},
		&domain.Prediction{},
		&domain.PredictionModification{},
		&domain.Migration{},
		&domain.SeedData{},
	}

	// Create migration record for auto-migration
	version := fmt.Sprintf("auto_%d", time.Now().Unix())
	migration := &domain.Migration{
		Version:    version,
		Name:       "GORM AutoMigrate",
		Type:       domain.MigrationTypeUp,
		Status:     domain.MigrationStatusRunning,
		ExecutedBy: "system",
	}

	migration.MarkAsRunning()
	if err := s.repository.CreateMigration(ctx, migration); err != nil {
		return fmt.Errorf("failed to create auto-migration record: %w", err)
	}

	start := time.Now()

	// Perform auto-migration
	err := s.db.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		migration.MarkAsFailed(err)
		s.repository.SaveMigration(ctx, migration)
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	// Mark as completed
	migration.MarkAsCompleted(time.Since(start))
	if err := s.repository.SaveMigration(ctx, migration); err != nil {
		s.logger.Warn("Failed to update auto-migration record: %v", err)
	}

	s.logger.Info("GORM auto-migration completed successfully in %v", time.Since(start))
	return nil
}

// RunMigrations executes all pending migrations from the migrations directory.
func (s *MigrationService) RunMigrations(ctx context.Context, migrationsDir string) error {
	s.logger.Info("Running manual migrations from directory: %s", migrationsDir)

	// Check for migration lock
	locked, err := s.repository.CheckMigrationLock(ctx)
	if err != nil {
		return fmt.Errorf("failed to check migration lock: %w", err)
	}
	if locked {
		return fmt.Errorf("migration is already running")
	}

	// Load migration files
	migrationFiles, err := s.loadMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	if len(migrationFiles) == 0 {
		s.logger.Info("No migration files found")
		return nil
	}

	// Get applied migrations
	appliedMigrations, err := s.repository.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Filter pending migrations
	var pendingMigrations []MigrationFile
	for _, file := range migrationFiles {
		if file.Type == domain.MigrationTypeUp && !appliedVersions[file.Version] {
			pendingMigrations = append(pendingMigrations, file)
		}
	}

	if len(pendingMigrations) == 0 {
		s.logger.Info("No pending migrations to run")
		return nil
	}

	s.logger.Info("Found %d pending migrations", len(pendingMigrations))

	// Execute pending migrations
	for _, migrationFile := range pendingMigrations {
		if err := s.executeMigration(ctx, migrationFile); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile.Version, err)
		}
	}

	s.logger.Info("All migrations completed successfully")
	return nil
}

// RollbackMigration rolls back the last applied migration.
func (s *MigrationService) RollbackMigration(ctx context.Context, migrationsDir string) error {
	s.logger.Info("Rolling back last migration...")

	// Get last applied migration
	lastMigration, err := s.repository.GetLastAppliedMigration(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	if !lastMigration.CanRollback() {
		return fmt.Errorf("migration %s cannot be rolled back", lastMigration.Version)
	}

	// Find corresponding down migration file
	downMigrationPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", lastMigration.Version, lastMigration.Name))
	
	if _, err := os.Stat(downMigrationPath); os.IsNotExist(err) {
		return fmt.Errorf("rollback file not found: %s", downMigrationPath)
	}

	// Load rollback migration
	content, err := os.ReadFile(downMigrationPath)
	if err != nil {
		return fmt.Errorf("failed to read rollback migration file: %w", err)
	}

	// Create rollback migration record
	rollbackMigration := &domain.Migration{
		Version:    lastMigration.Version + "_rollback",
		Name:       lastMigration.Name + "_rollback",
		Type:       domain.MigrationTypeDown,
		Status:     domain.MigrationStatusRunning,
		SQL:        string(content),
		Checksum:   fmt.Sprintf("%x", md5.Sum(content)),
		ExecutedBy: "system",
	}

	rollbackMigration.MarkAsRunning()
	if err := s.repository.CreateMigration(ctx, rollbackMigration); err != nil {
		return fmt.Errorf("failed to create rollback migration record: %w", err)
	}

	start := time.Now()

	// Execute rollback in transaction
	err = s.repository.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		// Execute rollback SQL
		if err := tx.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}

		// Mark original migration as rolled back
		lastMigration.MarkAsRolledBack()
		if err := tx.Save(lastMigration).Error; err != nil {
			return fmt.Errorf("failed to update original migration status: %w", err)
		}

		return nil
	})

	if err != nil {
		rollbackMigration.MarkAsFailed(err)
		s.repository.SaveMigration(ctx, rollbackMigration)
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Mark rollback as completed
	rollbackMigration.MarkAsCompleted(time.Since(start))
	if err := s.repository.SaveMigration(ctx, rollbackMigration); err != nil {
		s.logger.Warn("Failed to update rollback migration record: %v", err)
	}

	s.logger.Info("Migration %s rolled back successfully", lastMigration.Version)
	return nil
}

// GetMigrationStatus returns the current migration status.
func (s *MigrationService) GetMigrationStatus(ctx context.Context) (map[string]interface{}, error) {
	// Get database version
	version, err := s.repository.GetDatabaseVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database version: %w", err)
	}

	// Get migration counts
	appliedMigrations, err := s.repository.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	failedMigrations, err := s.repository.GetFailedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed migrations: %w", err)
	}

	// Check for lock
	locked, err := s.repository.CheckMigrationLock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check migration lock: %w", err)
	}

	return map[string]interface{}{
		"database_version":    version,
		"applied_migrations":  len(appliedMigrations),
		"failed_migrations":   len(failedMigrations),
		"migration_locked":    locked,
		"last_migration":      getLastMigrationInfo(appliedMigrations),
	}, nil
}

// RunSeedData executes seed data scripts.
func (s *MigrationService) RunSeedData(ctx context.Context, seedDataDir string) error {
	s.logger.Info("Running seed data from directory: %s", seedDataDir)

	// Load seed data files
	seedFiles, err := s.loadSeedDataFiles(seedDataDir)
	if err != nil {
		return fmt.Errorf("failed to load seed data files: %w", err)
	}

	if len(seedFiles) == 0 {
		s.logger.Info("No seed data files found")
		return nil
	}

	// Execute seed data
	for _, seedFile := range seedFiles {
		if err := s.executeSeedData(ctx, seedFile); err != nil {
			return fmt.Errorf("failed to execute seed data %s: %w", seedFile.Name, err)
		}
	}

	s.logger.Info("All seed data executed successfully")
	return nil
}

// ValidateMigrations validates migration files for consistency.
func (s *MigrationService) ValidateMigrations(ctx context.Context, migrationsDir string) error {
	s.logger.Info("Validating migrations...")

	migrationFiles, err := s.loadMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	// Check for missing down migrations
	upMigrations := make(map[string]bool)
	downMigrations := make(map[string]bool)

	for _, file := range migrationFiles {
		if file.Type == domain.MigrationTypeUp {
			upMigrations[file.Version] = true
		} else {
			downMigrations[file.Version] = true
		}
	}

	var missingDown []string
	for version := range upMigrations {
		if !downMigrations[version] {
			missingDown = append(missingDown, version)
		}
	}

	if len(missingDown) > 0 {
		s.logger.Warn("Missing down migrations for versions: %v", missingDown)
	}

	// Validate applied migrations against files
	appliedMigrations, err := s.repository.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, applied := range appliedMigrations {
		found := false
		for _, file := range migrationFiles {
			if file.Version == applied.Version && file.Type == domain.MigrationTypeUp {
				if file.Checksum != applied.Checksum {
					return fmt.Errorf("checksum mismatch for migration %s", applied.Version)
				}
				found = true
				break
			}
		}
		if !found {
			s.logger.Warn("Applied migration %s not found in migration files", applied.Version)
		}
	}

	s.logger.Info("Migration validation completed")
	return nil
}

// Helper methods

func (s *MigrationService) loadMigrationFiles(migrationsDir string) ([]MigrationFile, error) {
	var migrationFiles []MigrationFile

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		filename := d.Name()
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			return nil // Skip invalid filenames
		}

		version := parts[0]
		
		// Determine migration type
		var migrationType domain.MigrationType
		var name string
		
		if strings.HasSuffix(filename, ".up.sql") {
			migrationType = domain.MigrationTypeUp
			name = strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")
		} else if strings.HasSuffix(filename, ".down.sql") {
			migrationType = domain.MigrationTypeDown
			name = strings.TrimSuffix(strings.Join(parts[1:], "_"), ".down.sql")
		} else {
			return nil // Skip files that don't match pattern
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		migrationFiles = append(migrationFiles, MigrationFile{
			Version:  version,
			Name:     name,
			Type:     migrationType,
			FilePath: path,
			Content:  string(content),
			Checksum: fmt.Sprintf("%x", md5.Sum(content)),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by version
	sort.Slice(migrationFiles, func(i, j int) bool {
		vi, _ := strconv.Atoi(migrationFiles[i].Version)
		vj, _ := strconv.Atoi(migrationFiles[j].Version)
		return vi < vj
	})

	return migrationFiles, nil
}

func (s *MigrationService) executeMigration(ctx context.Context, migrationFile MigrationFile) error {
	s.logger.Info("Executing migration: %s - %s", migrationFile.Version, migrationFile.Name)

	// Create migration record
	migration := &domain.Migration{
		Version:    migrationFile.Version,
		Name:       migrationFile.Name,
		Type:       migrationFile.Type,
		Status:     domain.MigrationStatusRunning,
		SQL:        migrationFile.Content,
		Checksum:   migrationFile.Checksum,
		ExecutedBy: "system",
	}

	migration.MarkAsRunning()
	if err := s.repository.CreateMigration(ctx, migration); err != nil {
		return fmt.Errorf("failed to create migration record: %w", err)
	}

	start := time.Now()

	// Execute migration in transaction
	err := s.repository.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.Exec(migrationFile.Content).Error
	})

	if err != nil {
		migration.MarkAsFailed(err)
		s.repository.SaveMigration(ctx, migration)
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Mark as completed
	migration.MarkAsCompleted(time.Since(start))
	if err := s.repository.SaveMigration(ctx, migration); err != nil {
		s.logger.Warn("Failed to update migration record: %v", err)
	}

	s.logger.Info("Migration %s completed in %v", migrationFile.Version, time.Since(start))
	return nil
}

func (s *MigrationService) loadSeedDataFiles(seedDataDir string) ([]domain.SeedData, error) {
	// Implementation for loading seed data files
	// This is a placeholder - implement based on your seed data format
	return []domain.SeedData{}, nil
}

func (s *MigrationService) executeSeedData(ctx context.Context, seedData domain.SeedData) error {
	// Implementation for executing seed data
	// This is a placeholder - implement based on your seed data format
	return nil
}

func getLastMigrationInfo(migrations []domain.Migration) map[string]interface{} {
	if len(migrations) == 0 {
		return nil
	}

	last := migrations[len(migrations)-1]
	return map[string]interface{}{
		"version":     last.Version,
		"name":        last.Name,
		"executed_at": last.ExecutedAt,
		"duration":    last.Duration,
	}
}