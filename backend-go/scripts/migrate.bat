@echo off
REM Database Migration Script for Windows
REM This script provides convenient commands for database migration operations

setlocal enabledelayedexpansion

REM Configuration
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "MIGRATE_CMD=go run %PROJECT_ROOT%\cmd\migrate\main.go"
if "%CONFIG_FILE%"=="" set "CONFIG_FILE=config.development.yaml"
if "%MIGRATIONS_DIR%"=="" set "MIGRATIONS_DIR=migrations"
if "%SEED_DATA_DIR%"=="" set "SEED_DATA_DIR=seed_data"

REM Parse command line arguments
set "COMMAND="
set "FORCE=false"
set "VERBOSE=false"
set "TIMEOUT=30s"
set "MIGRATION_NAME="

:parse_args
if "%~1"=="" goto end_parse
if "%~1"=="--config" (
    set "CONFIG_FILE=%~2"
    shift
    shift
    goto parse_args
)
if "%~1"=="--force" (
    set "FORCE=true"
    shift
    goto parse_args
)
if "%~1"=="--verbose" (
    set "VERBOSE=true"
    shift
    goto parse_args
)
if "%~1"=="help" (
    goto show_help
)
if "%~1"=="up" (
    set "COMMAND=up"
    shift
    goto parse_args
)
if "%~1"=="down" (
    set "COMMAND=down"
    shift
    goto parse_args
)
if "%~1"=="status" (
    set "COMMAND=status"
    shift
    goto parse_args
)
if "%~1"=="validate" (
    set "COMMAND=validate"
    shift
    goto parse_args
)
if "%~1"=="seed" (
    set "COMMAND=seed"
    shift
    goto parse_args
)
if "%~1"=="auto" (
    set "COMMAND=auto"
    shift
    goto parse_args
)
if "%~1"=="init" (
    set "COMMAND=init"
    shift
    goto parse_args
)
if "%~1"=="create" (
    set "COMMAND=create"
    set "MIGRATION_NAME=%~2"
    shift
    shift
    goto parse_args
)
shift
goto parse_args

:end_parse

REM Set default command if none provided
if "%COMMAND%"=="" set "COMMAND=status"

REM Change to project root directory
cd /d "%PROJECT_ROOT%"

REM Handle commands
if "%COMMAND%"=="help" goto show_help
if "%COMMAND%"=="create" goto create_migration
if "%COMMAND%"=="up" goto run_up
if "%COMMAND%"=="down" goto run_down
if "%COMMAND%"=="status" goto run_status
if "%COMMAND%"=="validate" goto run_validate
if "%COMMAND%"=="seed" goto run_seed
if "%COMMAND%"=="auto" goto run_auto
if "%COMMAND%"=="init" goto run_init

echo [ERROR] Unknown command: %COMMAND%
goto show_help

:show_help
echo Database Migration Script for Windows
echo.
echo Usage: %~nx0 [COMMAND] [OPTIONS]
echo.
echo Commands:
echo     up          Run all pending migrations (auto + manual)
echo     down        Rollback the last migration (requires --force)
echo     status      Show current migration status
echo     validate    Validate migration files
echo     seed        Run seed data
echo     auto        Run GORM auto-migration only
echo     init        Initialize migration directory structure
echo     create      Create new migration files
echo     help        Show this help message
echo.
echo Options:
echo     --config FILE       Configuration file (default: config.development.yaml)
echo     --force             Force operation (required for rollbacks)
echo     --verbose           Enable verbose logging
echo.
echo Examples:
echo     %~nx0 up                           # Run all migrations
echo     %~nx0 down --force                 # Rollback last migration
echo     %~nx0 status                       # Check migration status
echo     %~nx0 create add_user_avatar       # Create new migration files
goto end

:create_migration
if "%MIGRATION_NAME%"=="" (
    echo [ERROR] Migration name is required
    echo Usage: %~nx0 create ^<migration_name^>
    exit /b 1
)

REM Generate timestamp-based version
for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set "DATE=%%d%%b%%c"
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set "TIME=%%a%%b"
set "VERSION=%DATE: =0%%TIME: =0%"
set "VERSION=%VERSION::=%"

set "UP_FILE=%MIGRATIONS_DIR%\%VERSION%_%MIGRATION_NAME%.up.sql"
set "DOWN_FILE=%MIGRATIONS_DIR%\%VERSION%_%MIGRATION_NAME%.down.sql"

REM Create migrations directory if it doesn't exist
if not exist "%MIGRATIONS_DIR%" mkdir "%MIGRATIONS_DIR%"

REM Create up migration file
(
echo -- Migration: %MIGRATION_NAME%
echo -- Created: %DATE% %TIME%
echo -- Description: Add description here
echo.
echo BEGIN;
echo.
echo -- Add your up migration SQL here
echo -- Example:
echo -- CREATE TABLE example ^(
echo --     id INT PRIMARY KEY AUTO_INCREMENT,
echo --     name VARCHAR^(255^) NOT NULL,
echo --     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
echo -- ^);
echo.
echo COMMIT;
) > "%UP_FILE%"

REM Create down migration file
(
echo -- Rollback: %MIGRATION_NAME%
echo -- Created: %DATE% %TIME%
echo -- Description: Rollback for %MIGRATION_NAME% migration
echo.
echo BEGIN;
echo.
echo -- Add your rollback SQL here
echo -- Example:
echo -- DROP TABLE IF EXISTS example;
echo.
echo COMMIT;
) > "%DOWN_FILE%"

echo [SUCCESS] Created migration files:
echo   Up:   %UP_FILE%
echo   Down: %DOWN_FILE%
echo [WARNING] Don't forget to add your SQL statements to both files!
goto end

:run_up
echo [INFO] Running all migrations (auto + manual)...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=up %VERBOSE_FLAG%
goto end

:run_down
if "%FORCE%"=="false" (
    echo [ERROR] Rollback requires --force flag for safety
    echo [WARNING] This will rollback the last migration and cannot be undone
    echo Use: %~nx0 down --force
    exit /b 1
)
echo [WARNING] Rolling back last migration...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=down --force %VERBOSE_FLAG%
goto end

:run_status
echo [INFO] Checking migration status...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=status %VERBOSE_FLAG%
goto end

:run_validate
echo [INFO] Validating migrations...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=validate %VERBOSE_FLAG%
goto end

:run_seed
echo [INFO] Running seed data...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=seed %VERBOSE_FLAG%
goto end

:run_auto
echo [INFO] Running GORM auto-migration...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=auto %VERBOSE_FLAG%
goto end

:run_init
echo [INFO] Initializing migration structure...
%MIGRATE_CMD% -config=%CONFIG_FILE% -migrations=%MIGRATIONS_DIR% -seed=%SEED_DATA_DIR% -command=init %VERBOSE_FLAG%
goto end

:end
echo [SUCCESS] Migration operation completed successfully!