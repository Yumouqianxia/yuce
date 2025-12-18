// Package domain contains the core business entities and domain logic.
//
// This file defines the migration system entities and interfaces for managing
// database schema changes in a structured and version-controlled manner.
package domain

import (
	"time"
)

// MigrationStatus represents the current state of a migration.
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "PENDING"   // Migration is ready to be applied
	MigrationStatusRunning   MigrationStatus = "RUNNING"   // Migration is currently being executed
	MigrationStatusCompleted MigrationStatus = "COMPLETED" // Migration has been successfully applied
	MigrationStatusFailed    MigrationStatus = "FAILED"    // Migration failed during execution
	MigrationStatusRolledBack MigrationStatus = "ROLLED_BACK" // Migration was rolled back
)

// MigrationType represents the type of migration operation.
type MigrationType string

const (
	MigrationTypeUp   MigrationType = "UP"   // Forward migration
	MigrationTypeDown MigrationType = "DOWN" // Rollback migration
)

// Migration represents a database migration record.
//
// This entity tracks the execution history of database migrations,
// providing audit trail and rollback capabilities.
type Migration struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Version     string          `gorm:"uniqueIndex;size:50;not null" json:"version"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Type        MigrationType   `gorm:"size:10;not null" json:"type"`
	Status      MigrationStatus `gorm:"size:20;not null;default:PENDING" json:"status"`
	SQL         string          `gorm:"type:text" json:"sql,omitempty"`
	Checksum    string          `gorm:"size:64" json:"checksum"`
	ExecutedBy  string          `gorm:"size:100" json:"executed_by"`
	ExecutedAt  *time.Time      `json:"executed_at"`
	Duration    int64           `json:"duration"` // Duration in milliseconds
	ErrorMsg    string          `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TableName returns the database table name for the Migration entity.
func (Migration) TableName() string {
	return "schema_migrations"
}

// IsCompleted returns true if the migration has been successfully applied.
func (m *Migration) IsCompleted() bool {
	return m.Status == MigrationStatusCompleted
}

// IsFailed returns true if the migration failed during execution.
func (m *Migration) IsFailed() bool {
	return m.Status == MigrationStatusFailed
}

// CanRollback returns true if the migration can be rolled back.
func (m *Migration) CanRollback() bool {
	return m.Status == MigrationStatusCompleted && m.Type == MigrationTypeUp
}

// MarkAsRunning marks the migration as currently running.
func (m *Migration) MarkAsRunning() {
	m.Status = MigrationStatusRunning
	now := time.Now()
	m.ExecutedAt = &now
}

// MarkAsCompleted marks the migration as successfully completed.
func (m *Migration) MarkAsCompleted(duration time.Duration) {
	m.Status = MigrationStatusCompleted
	m.Duration = duration.Milliseconds()
}

// MarkAsFailed marks the migration as failed with an error message.
func (m *Migration) MarkAsFailed(err error) {
	m.Status = MigrationStatusFailed
	if err != nil {
		m.ErrorMsg = err.Error()
	}
}

// MarkAsRolledBack marks the migration as rolled back.
func (m *Migration) MarkAsRolledBack() {
	m.Status = MigrationStatusRolledBack
}

// SeedData represents seed data for database initialization.
type SeedData struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Version     string    `gorm:"size:20;not null" json:"version"`
	Applied     bool      `gorm:"default:false" json:"applied"`
	AppliedAt   *time.Time `json:"applied_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the database table name for the SeedData entity.
func (SeedData) TableName() string {
	return "seed_data"
}

// MarkAsApplied marks the seed data as applied.
func (s *SeedData) MarkAsApplied() {
	s.Applied = true
	now := time.Now()
	s.AppliedAt = &now
}