// Package pkg contains reusable public packages for the prediction system.
//
// This directory houses packages that provide common functionality and can be
// imported by other projects. These packages are designed to be independent,
// well-documented, and follow Go best practices for public APIs.
//
// # Available Packages
//
// ## database
//
// The database package provides MySQL database connection management using GORM.
// It includes connection pooling, health monitoring, transaction support, and
// comprehensive logging capabilities.
//
// Key features:
//   - Automatic connection pooling with configurable parameters
//   - Health checks and connection monitoring
//   - Transaction support with context propagation
//   - Structured logging with different verbosity levels
//   - Graceful connection management and cleanup
//
// Example usage:
//
//	cfg := &config.DatabaseConfig{
//		Host:     "localhost",
//		Port:     3306,
//		Database: "myapp",
//		Username: "user",
//		Password: "pass",
//	}
//
//	db, err := database.NewDB(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
// ## redis
//
// The redis package provides Redis client management with support for both
// single-node and cluster configurations. It includes connection pooling,
// health monitoring, and metrics collection.
//
// Key features:
//   - Support for single-node and cluster Redis deployments
//   - Connection pooling with automatic reconnection
//   - Health monitoring and metrics collection
//   - Thread-safe operations with proper synchronization
//   - Graceful shutdown and resource cleanup
//
// Example usage:
//
//	cfg := &config.RedisConfig{
//		Host:     "localhost",
//		Port:     6379,
//		Database: 0,
//		PoolSize: 10,
//	}
//
//	err := redis.Initialize(cfg, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer redis.Shutdown()
//
// ## response
//
// The response package provides standardized HTTP response utilities for REST APIs.
// It ensures consistent response formats across all endpoints and integrates
// seamlessly with the Gin web framework.
//
// Key features:
//   - Standardized JSON response structure
//   - Type-safe error handling with detailed information
//   - Convenient helper functions for common HTTP status codes
//   - Integration with Gin framework
//   - Support for both success and error responses
//
// Example usage:
//
//	// Success response
//	response.OK(c, "User retrieved successfully", user)
//
//	// Error response
//	response.BadRequest(c, "Invalid user ID format")
//
// ## cache
//
// The cache package provides a layered caching system with support for both
// in-memory and Redis-based caching. It implements cache-aside and write-through
// patterns for optimal performance.
//
// Key features:
//   - Multi-layer caching (memory + Redis)
//   - Configurable TTL and eviction policies
//   - Cache statistics and monitoring
//   - Thread-safe operations
//   - Automatic cache warming and invalidation
//
// ## events
//
// The events package implements an event-driven architecture with support for
// both synchronous and asynchronous event processing. It provides a clean
// abstraction for decoupling components through events.
//
// Key features:
//   - Event bus with publish-subscribe pattern
//   - Support for both sync and async event handlers
//   - Event filtering and routing
//   - Error handling and retry mechanisms
//   - Event persistence and replay capabilities
//
// ## middleware
//
// The middleware package provides common HTTP middleware for Gin applications,
// including authentication, logging, error handling, and request tracking.
//
// Key features:
//   - JWT authentication middleware
//   - Request logging with correlation IDs
//   - Error handling and recovery
//   - CORS configuration
//   - Rate limiting and throttling
//
// # Design Principles
//
// All packages in this directory follow these design principles:
//
// ## Independence
//
// Each package is designed to be independent and can be used without requiring
// other packages from this project. Dependencies are kept minimal and well-defined.
//
// ## Documentation
//
// All public APIs are thoroughly documented with examples and usage patterns.
// Package documentation explains the purpose, key features, and common use cases.
//
// ## Testing
//
// Each package includes comprehensive unit tests with high coverage. Integration
// tests are provided where appropriate to test interactions with external systems.
//
// ## Error Handling
//
// All packages implement robust error handling with meaningful error messages
// and proper error wrapping for debugging and monitoring.
//
// ## Performance
//
// Packages are optimized for performance with minimal allocations, efficient
// algorithms, and proper resource management.
//
// ## Thread Safety
//
// All packages that may be used concurrently are designed to be thread-safe
// with proper synchronization mechanisms.
//
// # Usage Guidelines
//
// When using these packages:
//
// 1. Always check error returns and handle them appropriately
// 2. Use context.Context for cancellation and timeouts where supported
// 3. Follow the configuration patterns established by each package
// 4. Properly close/cleanup resources when done
// 5. Monitor metrics and health checks provided by the packages
//
// # Contributing
//
// When adding new packages to this directory:
//
// 1. Ensure the package provides general-purpose functionality
// 2. Write comprehensive documentation with examples
// 3. Include unit tests with good coverage
// 4. Follow Go naming conventions and best practices
// 5. Keep dependencies minimal and well-justified
// 6. Provide configuration options for flexibility
// 7. Include health checks and monitoring where appropriate
package pkg
