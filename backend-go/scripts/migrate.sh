#!/bin/bash

# Database Migration Script
# This script provides convenient commands for database migration operations

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATE_CMD="go run $PROJECT_ROOT/cmd/migrate/main.go"
CONFIG_FILE="${CONFIG_FILE:-config.development.yaml}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations}"
SEED_DATA_DIR="${SEED_DATA_DIR:-seed_data}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Database Migration Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    up          Run all pending migrations (auto + manual)
    down        Rollback the last migration (requires --force)
    status      Show current migration status
    validate    Validate migration files
    seed        Run seed data
    auto        Run GORM auto-migration only
    init        Initialize migration directory structure
    create      Create new migration files
    help        Show this help message

Options:
    --config=FILE       Configuration file (default: config.development.yaml)
    --migrations=DIR    Migrations directory (default: migrations)
    --seed=DIR          Seed data directory (default: seed_data)
    --force             Force operation (required for rollbacks)
    --verbose           Enable verbose logging
    --timeout=DURATION  Operation timeout (default: 30s)

Examples:
    $0 up                           # Run all migrations
    $0 down --force                 # Rollback last migration
    $0 status                       # Check migration status
    $0 create add_user_avatar       # Create new migration files
    $0 validate                     # Validate migrations
    $0 seed                         # Run seed data
    $0 auto                         # Run auto-migration only

Environment Variables:
    CONFIG_FILE         Override default config file
    MIGRATIONS_DIR      Override default migrations directory
    SEED_DATA_DIR       Override default seed data directory
    DB_HOST            Database host
    DB_PORT            Database port
    DB_NAME            Database name
    DB_USER            Database username
    DB_PASSWORD        Database password

EOF
}

create_migration() {
    local name="$1"
    if [ -z "$name" ]; then
        log_error "Migration name is required"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi

    # Generate timestamp-based version
    local version=$(date +"%Y%m%d%H%M%S")
    local up_file="$MIGRATIONS_DIR/${version}_${name}.up.sql"
    local down_file="$MIGRATIONS_DIR/${version}_${name}.down.sql"

    # Create migrations directory if it doesn't exist
    mkdir -p "$MIGRATIONS_DIR"

    # Create up migration file
    cat > "$up_file" << EOF
-- Migration: $name
-- Created: $(date)
-- Description: Add description here

BEGIN;

-- Add your up migration SQL here
-- Example:
-- CREATE TABLE example (
--     id INT PRIMARY KEY AUTO_INCREMENT,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

COMMIT;
EOF

    # Create down migration file
    cat > "$down_file" << EOF
-- Rollback: $name
-- Created: $(date)
-- Description: Rollback for $name migration

BEGIN;

-- Add your rollback SQL here
-- Example:
-- DROP TABLE IF EXISTS example;

COMMIT;
EOF

    log_success "Created migration files:"
    log_info "  Up:   $up_file"
    log_info "  Down: $down_file"
    log_warning "Don't forget to add your SQL statements to both files!"
}

run_migration_command() {
    local cmd="$1"
    shift
    local args=("$@")

    log_info "Running migration command: $cmd"
    
    # Build command arguments
    local migrate_args=(
        "-config=$CONFIG_FILE"
        "-migrations=$MIGRATIONS_DIR"
        "-seed=$SEED_DATA_DIR"
        "-command=$cmd"
    )

    # Add additional arguments
    for arg in "${args[@]}"; do
        migrate_args+=("$arg")
    done

    # Execute migration command
    if $MIGRATE_CMD "${migrate_args[@]}"; then
        log_success "Migration command '$cmd' completed successfully"
    else
        log_error "Migration command '$cmd' failed"
        exit 1
    fi
}

check_prerequisites() {
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check if config file exists
    if [ ! -f "$CONFIG_FILE" ]; then
        log_warning "Config file not found: $CONFIG_FILE"
        log_info "Make sure the config file exists or set CONFIG_FILE environment variable"
    fi

    # Check if we're in the project root
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        log_error "Not in Go project root directory"
        exit 1
    fi
}

# Parse command line arguments
COMMAND=""
FORCE=false
VERBOSE=false
TIMEOUT="30s"

while [[ $# -gt 0 ]]; do
    case $1 in
        --config=*)
            CONFIG_FILE="${1#*=}"
            shift
            ;;
        --migrations=*)
            MIGRATIONS_DIR="${1#*=}"
            shift
            ;;
        --seed=*)
            SEED_DATA_DIR="${1#*=}"
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --timeout=*)
            TIMEOUT="${1#*=}"
            shift
            ;;
        help|--help|-h)
            show_help
            exit 0
            ;;
        up|down|status|validate|seed|auto|init|create)
            COMMAND="$1"
            shift
            break
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Set default command if none provided
if [ -z "$COMMAND" ]; then
    COMMAND="status"
fi

# Check prerequisites
check_prerequisites

# Change to project root directory
cd "$PROJECT_ROOT"

# Handle special commands
case "$COMMAND" in
    help)
        show_help
        exit 0
        ;;
    create)
        create_migration "$1"
        exit 0
        ;;
esac

# Build additional arguments
ADDITIONAL_ARGS=()

if [ "$FORCE" = true ]; then
    ADDITIONAL_ARGS+=("--force")
fi

if [ "$VERBOSE" = true ]; then
    ADDITIONAL_ARGS+=("--verbose")
fi

ADDITIONAL_ARGS+=("--timeout=$TIMEOUT")

# Execute migration command
case "$COMMAND" in
    up)
        log_info "Running all migrations (auto + manual)..."
        run_migration_command "up" "${ADDITIONAL_ARGS[@]}"
        ;;
    down)
        if [ "$FORCE" != true ]; then
            log_error "Rollback requires --force flag for safety"
            log_warning "This will rollback the last migration and cannot be undone"
            log_info "Use: $0 down --force"
            exit 1
        fi
        log_warning "Rolling back last migration..."
        run_migration_command "down" "${ADDITIONAL_ARGS[@]}"
        ;;
    status)
        log_info "Checking migration status..."
        run_migration_command "status" "${ADDITIONAL_ARGS[@]}"
        ;;
    validate)
        log_info "Validating migrations..."
        run_migration_command "validate" "${ADDITIONAL_ARGS[@]}"
        ;;
    seed)
        log_info "Running seed data..."
        run_migration_command "seed" "${ADDITIONAL_ARGS[@]}"
        ;;
    auto)
        log_info "Running GORM auto-migration..."
        run_migration_command "auto" "${ADDITIONAL_ARGS[@]}"
        ;;
    init)
        log_info "Initializing migration structure..."
        run_migration_command "init" "${ADDITIONAL_ARGS[@]}"
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac

log_success "Migration operation completed successfully!"