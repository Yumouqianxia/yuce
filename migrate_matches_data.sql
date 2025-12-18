-- 数据迁移脚本：从 matches_backup 迁移数据到新的 matches 表
-- 清空当前的测试数据
DELETE FROM matches;

-- 重置自增ID
ALTER TABLE matches AUTO_INCREMENT = 1;

-- 迁移数据，字段映射：
-- optionA -> team_a
-- optionB -> team_b  
-- tournamentType -> tournament (spring/summer/winter)
-- matchTime -> start_time
-- status -> status (需要转换格式)
-- winner -> winner (A/B -> team_a/team_b)
-- scoreA -> score_a
-- scoreB -> score_b
-- createdAt -> created_at
-- updatedAt -> updated_at

INSERT INTO matches (
    id,
    team_a, 
    team_b,
    tournament,
    sport_type_id,
    status,
    start_time,
    winner,
    score_a,
    score_b,
    created_at,
    updated_at
)
SELECT 
    id,
    optionA as team_a,
    optionB as team_b,
    CASE 
        WHEN tournamentType = 'spring' THEN 'SPRING'
        WHEN tournamentType = 'summer' THEN 'SUMMER' 
        WHEN tournamentType = 'winter' THEN 'WINTER'
        ELSE 'SPRING'
    END as tournament,
    1 as sport_type_id, -- 默认设为1，可以后续调整
    CASE 
        WHEN status = 'not_started' THEN 'UPCOMING'
        WHEN status = 'in_progress' THEN 'LIVE'
        WHEN status = 'completed' THEN 'FINISHED'
        ELSE 'UPCOMING'
    END as status,
    STR_TO_DATE(matchTime, '%Y-%m-%d %H:%i:%s.%f') as start_time,
    CASE 
        WHEN winner = 'A' THEN 'team_a'
        WHEN winner = 'B' THEN 'team_b'
        ELSE NULL
    END as winner,
    scoreA as score_a,
    scoreB as score_b,
    createdAt as created_at,
    updatedAt as updated_at
FROM matches_backup
WHERE isActive = 1
ORDER BY id;

-- 检查迁移结果
SELECT COUNT(*) as migrated_matches FROM matches;
SELECT 
    tournament,
    status,
    COUNT(*) as count
FROM matches 
GROUP BY tournament, status
ORDER BY tournament, status;