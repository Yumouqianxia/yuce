# Database Migrations

This directory contains database migration files for the application.

## File Naming Convention

Migration files should follow this naming pattern:
- Up migrations: `{version}_{description}.up.sql`
- Down migrations: `{version}_{description}.down.sql`

Example:
- `001_create_users_table.up.sql`
- `001_create_users_table.down.sql`

## Usage

### Run all pending migrations:
```bash
./scripts/migrate.sh up
```

### Rollback last migration:
```bash
./scripts/migrate.sh down --force
```

### Check migration status:
```bash
./scripts/migrate.sh status
```

### Validate migrations:
```bash
./scripts/migrate.sh validate
```

### Run GORM auto-migration:
```bash
./scripts/migrate.sh auto
```

### Create new migration:
```bash
./scripts/migrate.sh create add_user_avatar
```

## Best Practices

1. **Always create both up and down migrations**
2. **Test migrations on a copy of production data**
3. **Keep migrations small and focused**
4. **Use transactions for complex migrations**
5. **Never modify existing migration files after they've been applied**
6. **Use descriptive names for migrations**
7. **Add comments to explain complex operations**

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

## Example Migration Files

### Up Migration (001_create_users_table.up.sql)
```sql
-- Migration: create_users_table
-- Created: 2023-12-01
-- Description: Create users table with basic fields

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    avatar VARCHAR(255),
    points INT DEFAULT 0,
    role ENUM('user', 'admin') DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_password_change DATETIME
);

-- Create indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_points ON users(points DESC);
CREATE INDEX idx_users_created_at ON users(created_at);

COMMIT;
```

### Down Migration (001_create_users_table.down.sql)
```sql
-- Rollback: create_users_table
-- Created: 2023-12-01
-- Description: Drop users table and related indexes

BEGIN;

-- Drop indexes first
DROP INDEX IF EXISTS idx_users_created_at ON users;
DROP INDEX IF EXISTS idx_users_points ON users;
DROP INDEX IF EXISTS idx_users_email ON users;
DROP INDEX IF EXISTS idx_users_username ON users;

-- Drop table
DROP TABLE IF EXISTS users;

COMMIT;
```

## Migration Status

You can check the current migration status using:

```bash
./scripts/migrate.sh status
```

This will show:
- Current database version
- Number of applied migrations
- Number of failed migrations
- Migration lock status
- Last migration details

## Troubleshooting

### Common Issues

1. **Migration Lock**: If a migration is stuck, it will be automatically cleared after 1 hour
2. **Checksum Mismatch**: Don't modify applied migration files; create new ones instead
3. **Missing Down Migration**: Always create both up and down migration files
4. **Database Connection**: Check your database configuration and connectivity

### Validation

Run migration validation to check for common issues:

```bash
./scripts/migrate.sh validate
```

This checks:
- File naming conventions
- Missing down migrations
- Checksum consistency
- Applied vs file migrations

## Production Deployment

1. **Backup** your database before running migrations
2. **Test** migrations on staging environment first
3. **Validate** all migration files
4. **Run** migrations during maintenance window
5. **Monitor** application after deployment
6. **Have rollback plan** ready

```bash
# Backup database
mysqldump -u root -p prediction_system > backup_$(date +%Y%m%d_%H%M%S).sql

# Run migrations
./scripts/migrate.sh up --config=config.production.yaml

# Verify status
./scripts/migrate.sh status --config=config.production.yaml
```