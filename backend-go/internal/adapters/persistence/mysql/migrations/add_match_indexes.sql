-- 为比赛表添加索引以优化查询性能

-- 比赛状态索引 (用于按状态查询)
CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status);

-- 比赛开始时间索引 (用于按时间排序和范围查询)
CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time);

-- 赛事类型索引 (用于按赛事查询)
CREATE INDEX IF NOT EXISTS idx_matches_tournament ON matches(tournament);

-- 复合索引：状态+开始时间 (用于获取特定状态的比赛并按时间排序)
CREATE INDEX IF NOT EXISTS idx_matches_status_start_time ON matches(status, start_time);

-- 复合索引：赛事+状态 (用于按赛事和状态查询)
CREATE INDEX IF NOT EXISTS idx_matches_tournament_status ON matches(tournament, status);

-- 复合索引：赛事+开始时间 (用于按赛事查询并按时间排序)
CREATE INDEX IF NOT EXISTS idx_matches_tournament_start_time ON matches(tournament, start_time);

-- 复合索引：状态+赛事+开始时间 (用于复杂查询)
CREATE INDEX IF NOT EXISTS idx_matches_status_tournament_start_time ON matches(status, tournament, start_time);

-- 创建时间索引 (用于按创建时间排序)
CREATE INDEX IF NOT EXISTS idx_matches_created_at ON matches(created_at);

-- 更新时间索引 (用于按更新时间排序)
CREATE INDEX IF NOT EXISTS idx_matches_updated_at ON matches(updated_at);