package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"gorm.io/gorm"
)

// ExampleNewDB demonstrates how to create a new database connection.
func ExampleNewDB() {
	// Create database configuration
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "prediction_system",
		Username:        "app_user",
		Password:        "secure_password",
		Charset:         "utf8mb4",
		Collation:       "utf8mb4_unicode_ci",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	// Create database connection
	db, err := NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Use the database
	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	fmt.Println("Database connection established successfully")
}

// ExampleDB_Transaction demonstrates how to use database transactions.
func ExampleDB_Transaction() {
	// Assume db is already initialized
	var db *DB

	// Define a transaction function
	err := db.Transaction(func(tx *gorm.DB) error {
		// Create a user
		user := &User{
			Username: "john_doe",
			Email:    "john@example.com",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Create a prediction for the user
		prediction := &Prediction{
			UserID:          user.ID,
			MatchID:         1,
			PredictedWinner: "A",
		}
		if err := tx.Create(prediction).Error; err != nil {
			return err
		}

		// Both operations succeed or both fail
		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}
}

// ExampleDB_HealthCheck demonstrates how to perform health checks.
func ExampleDB_HealthCheck() {
	// Assume db is already initialized
	var db *DB

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx); err != nil {
		log.Printf("Database is unhealthy: %v", err)
		return
	}

	// Get connection statistics
	stats := db.Stats()
	fmt.Printf("Database connections - Open: %d, In Use: %d, Idle: %d\n",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// Get connection info
	info := db.GetConnectionInfo()
	fmt.Printf("Database info: %+v\n", info)
}

// ExampleCreateDatabase demonstrates how to create a database if it doesn't exist.
func ExampleCreateDatabase() {
	cfg := &config.DatabaseConfig{
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "root_password",
		Database:  "new_database",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_unicode_ci",
	}

	// Create the database if it doesn't exist
	if err := CreateDatabase(cfg); err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	fmt.Println("Database created successfully")
}

// Example domain models for demonstration
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;size:50"`
	Email    string `gorm:"uniqueIndex;size:100"`
}

type Prediction struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"index"`
	MatchID         uint   `gorm:"index"`
	PredictedWinner string `gorm:"size:10"`
}
