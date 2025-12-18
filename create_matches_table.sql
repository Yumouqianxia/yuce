USE prediction_system;

-- Create matches table structure
CREATE TABLE matches (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    team_a VARCHAR(100) NOT NULL COMMENT 'Team A name',
    team_b VARCHAR(100) NOT NULL COMMENT 'Team B name',
    tournament VARCHAR(50) NOT NULL DEFAULT 'SPRING' COMMENT 'Tournament type',
    sport_type_id BIGINT UNSIGNED DEFAULT NULL COMMENT 'Sport type ID',
    status VARCHAR(20) NOT NULL DEFAULT 'UPCOMING' COMMENT 'Match status',
    start_time DATETIME NOT NULL COMMENT 'Start time',
    winner VARCHAR(10) DEFAULT NULL COMMENT 'Winner (A/B)',
    score_a INT NOT NULL DEFAULT 0 COMMENT 'Team A score',
    score_b INT NOT NULL DEFAULT 0 COMMENT 'Team B score',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Update time',
    
    INDEX idx_matches_status (status),
    INDEX idx_matches_start_time (start_time),
    INDEX idx_matches_tournament (tournament),
    INDEX idx_matches_status_start_time (status, start_time),
    INDEX idx_matches_sport_type (sport_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Matches table';

-- Add constraint checks
ALTER TABLE matches 
ADD CONSTRAINT chk_matches_status 
CHECK (status IN ('UPCOMING', 'LIVE', 'FINISHED', 'CANCELLED'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_tournament 
CHECK (tournament IN ('SPRING', 'SUMMER', 'WORLDS'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_winner 
CHECK (winner IS NULL OR winner IN ('A', 'B'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_scores 
CHECK (score_a >= 0 AND score_b >= 0);

-- Insert test data
INSERT INTO matches (team_a, team_b, tournament, status, start_time) VALUES
('T1', 'GenG', 'WORLDS', 'UPCOMING', '2025-09-07 20:00:00'),
('DK', 'KT', 'SPRING', 'FINISHED', '2025-09-06 18:00:00'),
('BRO', 'HLE', 'SUMMER', 'LIVE', '2025-09-06 22:00:00');

-- Update results for finished matches
UPDATE matches SET winner = 'A', score_a = 2, score_b = 1 WHERE team_a = 'DK' AND team_b = 'KT';