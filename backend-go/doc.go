// Package backend-go implements a high-performance prediction system backend.
//
// This application is built using Go with a focus on performance, scalability,
// and maintainability. It implements a modern hexagonal architecture pattern
// with clean separation of concerns between business logic and infrastructure.
//
// # Architecture Overview
//
// The system follows hexagonal architecture (ports and adapters) principles:
//
//   - Core Domain: Business logic and entities (internal/core/domain)
//   - Application Services: Use cases and orchestration (internal/core/services)
//   - Ports: Interfaces defining contracts (internal/core/ports)
//   - Adapters: Infrastructure implementations (internal/adapters)
//   - Shared Packages: Reusable utilities (pkg/)
//
// # Technology Stack
//
//   - Framework: Gin (HTTP routing and middleware)
//   - Database: MySQL 8.0 with GORM ORM
//   - Cache: Redis 6.0+ for high-performance caching
//   - Authentication: JWT tokens with bcrypt password hashing
//   - Real-time: WebSocket for live updates
//   - Configuration: Viper for flexible configuration management
//   - Logging: Structured logging with Logrus
//
// # Key Features
//
//   - Sports match prediction system
//   - User voting and ranking system
//   - Real-time updates via WebSocket
//   - Comprehensive caching strategy
//   - Event-driven architecture
//   - Robust error handling and logging
//   - Health monitoring and metrics
//   - Docker containerization support
//
// # Performance Characteristics
//
// The system is optimized for a 2C4G cloud server environment:
//
//   - Memory usage: <500MB total (Go app + MySQL + Redis)
//   - Concurrent connections: 2000+ WebSocket connections
//   - API response time: <100ms for most endpoints
//   - Database connection pool: Optimized for 20 max connections
//   - Cache hit ratio: >90% for frequently accessed data
//
// # Package Organization
//
// The codebase is organized into several key packages:
//
//   - cmd/: Application entry points and main functions
//   - internal/: Private application code not intended for external use
//   - pkg/: Public packages that can be imported by other projects
//   - api/: API definitions and documentation
//   - migrations/: Database schema migrations
//   - deployments/: Docker and deployment configurations
//   - docs/: Project documentation and guides
//
// # Getting Started
//
// To run the application:
//
//	# Install dependencies
//	go mod tidy
//
//	# Set up configuration
//	cp .env.example .env
//	# Edit .env with your database and Redis settings
//
//	# Run database migrations
//	make migrate
//
//	# Start the development server
//	make dev
//
// # Configuration
//
// The application uses environment-based configuration with support for:
//
//   - Environment variables
//   - Configuration files (YAML/JSON)
//   - Command-line flags
//   - Default values with validation
//
// # Testing
//
// The project includes comprehensive testing:
//
//	# Run all tests
//	make test
//
//	# Run tests with coverage
//	make test-coverage
//
//	# Run integration tests
//	make test-integration
//
// # Code Quality
//
// Code quality is maintained through:
//
//   - gofmt: Automatic code formatting
//   - staticcheck: Static analysis for bug detection
//   - golangci-lint: Comprehensive linting with multiple analyzers
//   - Unit tests: >80% code coverage target
//   - Integration tests: End-to-end API testing
//   - Pre-commit hooks: Automated quality checks
//
// # Monitoring and Observability
//
// The application provides comprehensive monitoring:
//
//   - Health check endpoints for all services
//   - Structured logging with correlation IDs
//   - Metrics collection for performance monitoring
//   - Database connection pool monitoring
//   - Redis cache performance metrics
//   - WebSocket connection tracking
//
// # Security
//
// Security features include:
//
//   - JWT-based authentication with refresh tokens
//   - bcrypt password hashing with configurable cost
//   - Input validation and sanitization
//   - SQL injection prevention via ORM
//   - Rate limiting and request throttling
//   - CORS configuration for cross-origin requests
//   - Secure headers and HTTPS enforcement
//
// # Deployment
//
// The application supports multiple deployment options:
//
//   - Docker containers with multi-stage builds
//   - Docker Compose for development environments
//   - Kubernetes manifests for production deployment
//   - Health checks and graceful shutdown
//   - Environment-specific configuration
//
// For detailed documentation, see the docs/ directory or visit the project wiki.
package main
