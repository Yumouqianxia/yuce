# Seed Data

This directory contains seed data files for initializing the database with default or test data.

## File Naming Convention

Seed data files should follow this naming pattern:
- `{version}_{description}.sql`

Example:
- `001_initial_users.sql`
- `002_sample_matches.sql`
- `003_test_predictions.sql`

## Usage

### Run all seed data:
```bash
./scripts/migrate.sh seed
```

### Run with specific configuration:
```bash
./scripts/migrate.sh seed --config=config.development.yaml
```

## Seed Data Types

### 1. Initial System Data
Essential data required for the application to function:
- Admin users
- System configurations
- Default roles and permissions

### 2. Development Data
Sample data for development and testing:
- Test users
- Sample matches
- Example predictions

### 3. Production Data
Minimal data required for production deployment:
- Admin accounts
- System settings

## Example Seed Files

### 001_initial_users.sql
```sql
-- Initial system users
-- This seed creates the default admin user and test users

INSERT INTO users (username, email, password, nickname, role, points, created_at, updated_at) VALUES
-- Admin user (password: admin123)
('admin', 'admin@prediction-system.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'System Admin', 'admin', 0, NOW(), NOW()),

-- Test users for development (password: password123)
('testuser1', 'test1@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test User 1', 'user', 100, NOW(), NOW()),
('testuser2', 'test2@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test User 2', 'user', 150, NOW(), NOW()),
('testuser3', 'test3@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test User 3', 'user', 200, NOW(), NOW());
```

### 002_sample_matches.sql
```sql
-- Sample matches for development and testing
-- This seed creates example matches for different tournaments

INSERT INTO matches (team_a, team_b, tournament, status, start_time, created_at, updated_at) VALUES
-- Upcoming matches
('Team Alpha', 'Team Beta', 'SPRING', 'UPCOMING', DATE_ADD(NOW(), INTERVAL 1 DAY), NOW(), NOW()),
('Team Gamma', 'Team Delta', 'SPRING', 'UPCOMING', DATE_ADD(NOW(), INTERVAL 2 DAY), NOW(), NOW()),
('Team Echo', 'Team Foxtrot', 'SUMMER', 'UPCOMING', DATE_ADD(NOW(), INTERVAL 3 DAY), NOW(), NOW()),

-- Live matches
('Team Golf', 'Team Hotel', 'SPRING', 'LIVE', DATE_SUB(NOW(), INTERVAL 1 HOUR), NOW(), NOW()),

-- Finished matches with results
('Team India', 'Team Juliet', 'SPRING', 'FINISHED', DATE_SUB(NOW(), INTERVAL 1 DAY), NOW(), NOW()),
('Team Kilo', 'Team Lima', 'SUMMER', 'FINISHED', DATE_SUB(NOW(), INTERVAL 2 DAY), NOW(), NOW()),
('Team Mike', 'Team November', 'WORLDS', 'FINISHED', DATE_SUB(NOW(), INTERVAL 3 DAY), NOW(), NOW());

-- Update finished matches with scores and winners
UPDATE matches SET score_a = 2, score_b = 1, winner = 'A' WHERE team_a = 'Team India' AND team_b = 'Team Juliet';
UPDATE matches SET score_a = 0, score_b = 3, winner = 'B' WHERE team_a = 'Team Kilo' AND team_b = 'Team Lima';
UPDATE matches SET score_a = 1, score_b = 1, winner = 'DRAW' WHERE team_a = 'Team Mike' AND team_b = 'Team November';
```

### 003_sample_predictions.sql
```sql
-- Sample predictions for development and testing
-- This seed creates example predictions for the sample matches

-- Get match IDs for predictions
SET @match1_id = (SELECT id FROM matches WHERE team_a = 'Team Alpha' AND team_b = 'Team Beta' LIMIT 1);
SET @match2_id = (SELECT id FROM matches WHERE team_a = 'Team Gamma' AND team_b = 'Team Delta' LIMIT 1);
SET @match3_id = (SELECT id FROM matches WHERE team_a = 'Team Echo' AND team_b = 'Team Foxtrot' LIMIT 1);
SET @finished_match_id = (SELECT id FROM matches WHERE team_a = 'Team India' AND team_b = 'Team Juliet' LIMIT 1);

-- Get user IDs for predictions
SET @user1_id = (SELECT id FROM users WHERE username = 'testuser1' LIMIT 1);
SET @user2_id = (SELECT id FROM users WHERE username = 'testuser2' LIMIT 1);
SET @user3_id = (SELECT id FROM users WHERE username = 'testuser3' LIMIT 1);

-- Insert sample predictions
INSERT INTO predictions (user_id, match_id, predicted_winner, predicted_score_a, predicted_score_b, created_at, updated_at) VALUES
-- Predictions for upcoming matches
(@user1_id, @match1_id, 'A', 2, 1, NOW(), NOW()),
(@user2_id, @match1_id, 'B', 1, 2, NOW(), NOW()),
(@user3_id, @match1_id, 'A', 3, 0, NOW(), NOW()),

(@user1_id, @match2_id, 'B', 0, 2, NOW(), NOW()),
(@user2_id, @match2_id, 'A', 2, 1, NOW(), NOW()),

(@user1_id, @match3_id, 'A', 1, 0, NOW(), NOW()),

-- Predictions for finished match (for scoring calculation)
(@user1_id, @finished_match_id, 'A', 2, 1, NOW(), NOW()), -- Correct prediction
(@user2_id, @finished_match_id, 'B', 1, 2, NOW(), NOW()), -- Wrong winner
(@user3_id, @finished_match_id, 'A', 1, 0, NOW(), NOW()); -- Correct winner, wrong score

-- Update predictions for finished match with calculated points
UPDATE predictions SET is_correct = TRUE, earned_points = 30 WHERE user_id = @user1_id AND match_id = @finished_match_id; -- Exact match
UPDATE predictions SET is_correct = FALSE, earned_points = 0 WHERE user_id = @user2_id AND match_id = @finished_match_id; -- Wrong winner
UPDATE predictions SET is_correct = TRUE, earned_points = 10 WHERE user_id = @user3_id AND match_id = @finished_match_id; -- Correct winner only
```

### 004_system_config.sql
```sql
-- System configuration data
-- This seed creates system-wide configuration settings

-- Note: This is an example - adjust based on your actual configuration table structure
-- INSERT INTO system_config (key, value, description, created_at, updated_at) VALUES
-- ('max_predictions_per_user', '10', 'Maximum number of predictions a user can make per match', NOW(), NOW()),
-- ('prediction_deadline_hours', '2', 'Hours before match start when predictions are locked', NOW(), NOW()),
-- ('points_correct_winner', '10', 'Points awarded for predicting correct winner', NOW(), NOW()),
-- ('points_exact_score', '20', 'Additional points for predicting exact score', NOW(), NOW()),
-- ('points_modification_penalty', '2', 'Points deducted per prediction modification', NOW(), NOW());
```

## Environment-Specific Seed Data

### Development Environment
- Includes test users with known passwords
- Sample matches and predictions
- Debug data for testing features

### Testing Environment
- Minimal test data for automated tests
- Consistent data for reproducible tests
- No sensitive information

### Production Environment
- Only essential system data
- Admin accounts with secure passwords
- No test or sample data

## Best Practices

1. **Use transactions** for related data
2. **Check for existing data** before inserting
3. **Use meaningful test data** that represents real scenarios
4. **Document passwords** for test accounts
5. **Keep production seeds minimal** and secure
6. **Version your seed data** like migrations
7. **Test seed data** in different environments

## Security Considerations

### Password Hashing
All passwords in seed data should be properly hashed using bcrypt:

```sql
-- Use bcrypt hashed passwords (example hash for 'password123')
'$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'
```

### Production Data
- Never include real user data in seeds
- Use secure passwords for admin accounts
- Remove or disable test accounts in production
- Audit seed data for sensitive information

## Running Seed Data

### Development
```bash
# Run all seed data for development
./scripts/migrate.sh seed --config=config.development.yaml
```

### Testing
```bash
# Run minimal seed data for testing
./scripts/migrate.sh seed --config=config.testing.yaml
```

### Production
```bash
# Run only essential seed data for production
./scripts/migrate.sh seed --config=config.production.yaml
```

## Troubleshooting

### Common Issues

1. **Duplicate Key Errors**: Check if data already exists
2. **Foreign Key Constraints**: Ensure referenced data exists
3. **Data Type Mismatches**: Verify column types and formats
4. **Permission Issues**: Check database user permissions

### Debugging

```bash
# Validate seed data syntax
mysql -u root -p --execute="source seed_data/001_initial_users.sql" prediction_system

# Check seed data status
./scripts/migrate.sh status
```

## Integration with Tests

Seed data can be used in automated tests:

```go
func TestWithSeedData(t *testing.T) {
    // Seed data provides known test users
    testUser := &User{
        Username: "testuser1",
        Email:    "test1@example.com",
    }
    
    // Test with seeded data
    user, err := userService.GetByUsername("testuser1")
    assert.NoError(t, err)
    assert.Equal(t, "Test User 1", user.Nickname)
}
```