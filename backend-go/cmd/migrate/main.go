// Package main provides the migration command-line tool.
//
// This tool provides comprehensive database migration management capabilities
// including GORM auto-migration, manual SQL migrations, rollbacks, and seed data.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backend-go/internal/adapters/persistence/mysql"
	"backend-go/internal/config"
	"backend-go/internal/core/services"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/database"
)

const (
	defaultMigrationsDir = "migrations"
	defaultSeedDataDir   = "seed_data"
	defaultTimeout       = 30 * time.Second
)

func main() {
	var (
		configPath    = flag.String("config", "config.development.yaml", "Path to configuration file")
		command       = flag.String("command", "up", "Migration command: up, down, status, validate, seed, auto")
		migrationsDir = flag.String("migrations", defaultMigrationsDir, "Path to migrations directory")
		seedDataDir   = flag.String("seed", defaultSeedDataDir, "Path to seed data directory")
		timeout       = flag.Duration("timeout", defaultTimeout, "Operation timeout")
		force         = flag.Bool("force", false, "Force operation (use with caution)")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Initialize logger
	if *verbose {
		logger.Init("debug")
	} else {
		logger.Init("info")
	}

	log := logger.GetLogger()
	log.Info("Starting migration tool...")

	// Load configuration
	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migration repository and service
	migrationRepo := mysql.NewMigrationRepository(db)
	migrationService := services.NewMigrationService(db, migrationRepo)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Initialize migration system
	if err := migrationService.InitializeMigrationSystem(ctx); err != nil {
		log.Fatal("Failed to initialize migration system: %v", err)
	}

	// Execute command
	switch *command {
	case "up":
		if err := runUpMigrations(ctx, migrationService, *migrationsDir); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "down":
		if err := runDownMigration(ctx, migrationService, *migrationsDir, *force); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
	case "status":
		if err := showMigrationStatus(ctx, migrationService); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
	case "validate":
		if err := validateMigrations(ctx, migrationService, *migrationsDir); err != nil {
			log.Fatalf("Migration validation failed: %v", err)
		}
	case "seed":
		if err := runSeedData(ctx, migrationService, *seedDataDir); err != nil {
			log.Fatalf("Seed data failed: %v", err)
		}
	case "auto":
		if err := runAutoMigration(ctx, migrationService); err != nil {
			log.Fatalf("Auto-migration failed: %v", err)
		}
	case "init":
		if err := initializeMigrationStructure(*migrationsDir, *seedDataDir); err != nil {
			log.Fatalf("Failed to initialize migration structure: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s. Available commands: up, down, status, validate, seed, auto, init", *command)
	}

	log.Info("Migration tool completed successfully")
}

func runUpMigrations(ctx context.Context, service *services.MigrationService, migrationsDir string) error {
	log := logger.GetLogger()
	log.Info("Running up migrations...")

	// First run auto-migration to ensure GORM models are up to date
	if err := service.AutoMigrate(ctx); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	// Then run manual migrations
	if err := service.RunMigrations(ctx, migrationsDir); err != nil {
		return fmt.Errorf("manual migrations failed: %w", err)
	}

	log.Info("All migrations completed successfully")
	return nil
}

func runDownMigration(ctx context.Context, service *services.MigrationService, migrationsDir string, force bool) error {
	log := logger.GetLogger()

	if !force {
		log.Warn("Rollback operation will revert the last migration. This action cannot be undone.")
		log.Warn("Use --force flag to confirm this operation.")
		return fmt.Errorf("rollback requires --force flag for safety")
	}

	log.Info("Rolling back last migration...")
	return service.RollbackMigration(ctx, migrationsDir)
}

func showMigrationStatus(ctx context.Context, service *services.MigrationService) error {
	log := logger.GetLogger()
	log.Info("Getting migration status...")

	status, err := service.GetMigrationStatus(ctx)
	if err != nil {
		return err
	}

	fmt.Println("\n=== Migration Status ===")
	fmt.Printf("Database Version: %v\n", status["database_version"])
	fmt.Printf("Applied Migrations: %v\n", status["applied_migrations"])
	fmt.Printf("Failed Migrations: %v\n", status["failed_migrations"])
	fmt.Printf("Migration Locked: %v\n", status["migration_locked"])

	if lastMigration := status["last_migration"]; lastMigration != nil {
		if migrationInfo, ok := lastMigration.(map[string]interface{}); ok {
			fmt.Printf("\nLast Migration:\n")
			fmt.Printf("  Version: %v\n", migrationInfo["version"])
			fmt.Printf("  Name: %v\n", migrationInfo["name"])
			fmt.Printf("  Executed At: %v\n", migrationInfo["executed_at"])
			fmt.Printf("  Duration: %v ms\n", migrationInfo["duration"])
		}
	}

	return nil
}

func validateMigrations(ctx context.Context, service *services.MigrationService, migrationsDir string) error {
	log := logger.GetLogger()
	log.Info("Validating migrations...")

	return service.ValidateMigrations(ctx, migrationsDir)
}

func runSeedData(ctx context.Context, service *services.MigrationService, seedDataDir string) error {
	log := logger.GetLogger()
	log.Info("Running seed data...")

	return service.RunSeedData(ctx, seedDataDir)
}

func runAutoMigration(ctx context.Context, service *services.MigrationService) error {
	log := logger.GetLogger()
	log.Info("Running GORM auto-migration...")

	return service.AutoMigrate(ctx)
}

func initializeMigrationStructure(migrationsDir, seedDataDir string) error {
	log := logger.GetLogger()
	log.Info("Initializing migration directory structure...")

	// Create migrations directory
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create seed data directory
	if err := os.MkdirAll(seedDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create seed data directory: %w", err)
	}

	// Create example migration files
	exampleUpMigration := `-- Example up migration
-- This file demonstrates the structure of an up migration

CREATE TABLE IF NOT EXISTS example_table (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_example_name ON example_table(name);
`

	exampleDownMigration := `-- Example down migration
-- This file demonstrates the structure of a down migration

DROP INDEX IF EXISTS idx_example_name ON example_table;
DROP TABLE IF EXISTS example_table;
`

	// Write example migration files
	upFile := filepath.Join(migrationsDir, "001_example_migration.up.sql")
	downFile := filepath.Join(migrationsDir, "001_example_migration.down.sql")

	if err := os.WriteFile(upFile, []byte(exampleUpMigration), 0644); err != nil {
		return fmt.Errorf("failed to create example up migration: %w", err)
	}

	if err := os.WriteFile(downFile, []byte(exampleDownMigration), 0644); err != nil {
		return fmt.Errorf("failed to create example down migration: %w", err)
	}

	// Create README file
	readme := `# Database Migrations

This directory contains database migration files for the application.

## File Naming Convention

Migration files should follow this naming pattern:
- Up migrations: {version}_{description}.up.sql
- Down migrations: {version}_{description}.down.sql

Example:
- 001_create_users_table.up.sql
- 001_create_users_table.down.sql

## Usage

### Run all pending migrations:
` + "```bash" + `
go run cmd/migrate/main.go -command=up
` + "```" + `

### Rollback last migration:
` + "```bash" + `
go run cmd/migrate/main.go -command=down --force
` + "```" + `

### Check migration status:
` + "```bash" + `
go run cmd/migrate/main.go -command=status
` + "```" + `

### Validate migrations:
` + "```bash" + `
go run cmd/migrate/main.go -command=validate
` + "```" + `

### Run GORM auto-migration:
` + "```bash" + `
go run cmd/migrate/main.go -command=auto
` + "```" + `

## Best Practices

1. Always create both up and down migrations
2. Test migrations on a copy of production data
3. Keep migrations small and focused
4. Use transactions for complex migrations
5. Never modify existing migration files after they've been applied
6. Use descriptive names for migrations
7. Add comments to explain complex operations

## Migration Types

### GORM Auto-Migration
- Automatically creates/updates tables based on Go struct definitions
- Good for development and basic schema changes
- Limited rollback capabilities

### Manual SQL Migrations
- Full control over database schema changes
- Support for complex operations and data transformations
- Complete rollback support with down migrations
- Required for production environments
`

	readmeFile := filepath.Join(migrationsDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to create README file: %w", err)
	}

	log.Info("Migration structure initialized successfully")
	log.Info("Created directories:")
	log.Info("  - %s (migrations)", migrationsDir)
	log.Info("  - %s (seed data)", seedDataDir)
	log.Info("Created example files:")
	log.Info("  - %s", upFile)
	log.Info("  - %s", downFile)
	log.Info("  - %s", readmeFile)

	return nil
}
