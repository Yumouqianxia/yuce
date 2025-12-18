package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/user"
	"backend-go/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "test":
		testDatabase()
	case "benchmark":
		benchmarkDatabase()
	case "stress":
		stressTest()
	case "migrate":
		runMigration()
	case "validate":
		validateDatabase()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Database Test Tool")
	fmt.Println("Usage: go run main.go <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  test      - Run basic database tests")
	fmt.Println("  benchmark - Run performance benchmarks")
	fmt.Println("  stress    - Run stress tests")
	fmt.Println("  migrate   - Run database migrations")
	fmt.Println("  validate  - Validate database configuration")
}

func testDatabase() {
	fmt.Println("Starting database tests...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Test configuration validation
	fmt.Println("Testing configuration validation...")
	if err := database.ValidateConfig(&cfg.Database); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
	fmt.Println("Configuration validation passed")

	// Test connection
	fmt.Println("Testing database connection...")
	if err := database.TestConnection(&cfg.Database); err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}
	fmt.Println("Connection test passed")

	// Initialize database
	fmt.Println("Initializing database...")
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	fmt.Println("Database initialized successfully")

	// Get database instance
	db := database.GetDB()
	if db == nil {
		log.Fatal("Failed to get database instance")
	}

	ctx := context.Background()

	// Test basic operations
	fmt.Println("Testing basic CRUD operations...")

	// Test create
	testUser := &user.User{
		Username: "test_user",
		Email:    "test@example.com",
	}

	err = database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
		return db.WithContext(ctx).Create(testUser).Error
	})
	if err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}
	fmt.Printf("Created user with ID: %d\n", testUser.ID)

	// Test query
	var foundUser user.User
	err = database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
		return db.WithContext(ctx).First(&foundUser, "username = ?", "test_user").Error
	})
	if err != nil {
		log.Fatalf("Failed to query user: %v", err)
	}
	fmt.Printf("Found user: %s (%s)\n", foundUser.Username, foundUser.Email)

	// Test update
	err = database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
		return db.WithContext(ctx).Model(&foundUser).Update("points", 100).Error
	})
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Println("Updated user points")

	// Test delete
	err = database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
		return db.WithContext(ctx).Delete(&foundUser).Error
	})
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Println("Deleted test user")

	fmt.Println("Basic CRUD tests completed successfully")
}

func benchmarkDatabase() {
	fmt.Println("Starting database benchmarks...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	ctx := context.Background()

	// Benchmark single inserts
	fmt.Println("Single insert benchmark...")
	start := time.Now()
	for i := 0; i < 100; i++ {
		testUser := &user.User{
			Username: fmt.Sprintf("bench_user_%d", i),
			Email:    fmt.Sprintf("bench_%d@example.com", i),
		}

		err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
			return db.WithContext(ctx).Create(testUser).Error
		})
		if err != nil {
			log.Printf("Failed to create user %d: %v", i, err)
		}
	}
	fmt.Printf("Single inserts completed in: %v\n", time.Since(start))

	// Benchmark batch inserts
	fmt.Println("Batch insert benchmark...")
	users := make([]user.User, 1000)
	for i := range users {
		users[i] = user.User{
			Username: fmt.Sprintf("batch_user_%d", i),
			Email:    fmt.Sprintf("batch_%d@example.com", i),
		}
	}

	start = time.Now()
	err = database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
		return db.WithContext(ctx).CreateInBatches(users, 100).Error
	})
	if err != nil {
		log.Fatalf("Batch insert failed: %v", err)
	}
	fmt.Printf("Batch insert completed in: %v\n", time.Since(start))

	// Benchmark queries
	fmt.Println("Query benchmark...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		var users []user.User
		err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
			return db.WithContext(ctx).Limit(10).Find(&users).Error
		})
		if err != nil {
			log.Printf("Query %d failed: %v", i, err)
		}
	}
	fmt.Printf("Queries completed in: %v\n", time.Since(start))

	fmt.Println("Benchmarks completed")
}

func stressTest() {
	fmt.Println("Starting stress test...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	ctx := context.Background()
	operations := 0
	errors := 0
	start := time.Now()

	// Run for 30 seconds
	for time.Since(start) < 30*time.Second {
		operations++

		switch operations % 4 {
		case 0: // Insert
			testUser := &user.User{
				Username: fmt.Sprintf("stress_user_%d", operations),
				Email:    fmt.Sprintf("stress_%d@example.com", operations),
			}

			err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
				return db.WithContext(ctx).Create(testUser).Error
			})
			if err != nil {
				errors++
			}

		case 1: // Query
			var users []user.User
			err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
				return db.WithContext(ctx).Limit(5).Find(&users).Error
			})
			if err != nil {
				errors++
			}

		case 2: // Update
			err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
				return db.WithContext(ctx).Model(&user.User{}).
					Where("username LIKE ?", "stress_user_%").
					Update("points", operations%1000).Error
			})
			if err != nil {
				errors++
			}

		case 3: // Count
			err := database.WithTransaction(ctx, func(db *database.DB) error {
				var count int64
				return db.WithContext(ctx).Model(&user.User{}).Count(&count).Error
			})
			if err != nil {
				errors++
			}
		}

		if operations%1000 == 0 {
			fmt.Printf("Operations: %d, Errors: %d\n", operations, errors)
		}
	}

	duration := time.Since(start)
	fmt.Printf("Stress test completed:\n")
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Printf("  Operations: %d\n", operations)
	fmt.Printf("  Errors: %d\n", errors)
	fmt.Printf("  Operations/sec: %.2f\n", float64(operations)/duration.Seconds())
	fmt.Printf("  Error rate: %.2f%%\n", float64(errors)/float64(operations)*100)
}

func runMigration() {
	fmt.Println("Running database migration...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	db := database.GetDB()
	if db == nil {
		log.Fatal("Failed to get database instance")
	}

	fmt.Println("Running auto migration...")
	err = db.AutoMigrate(
		&user.User{},
		&domain.Match{},
		&domain.Prediction{},
		&domain.PredictionModification{},
	)
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully")
}

func validateDatabase() {
	fmt.Println("Validating database configuration...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.ValidateConfig(&cfg.Database); err != nil {
		log.Fatalf("Database configuration is invalid: %v", err)
	}

	fmt.Println("Database configuration is valid")
}