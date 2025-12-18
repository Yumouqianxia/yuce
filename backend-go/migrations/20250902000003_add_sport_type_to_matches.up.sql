-- 为比赛表添加运动类型关联
ALTER TABLE matches 
ADD COLUMN sport_type_id BIGINT UNSIGNED DEFAULT NULL COMMENT '运动类型ID' AFTER tournament;

-- 添加外键约束
ALTER TABLE matches 
ADD CONSTRAINT fk_matches_sport_type 
FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE SET NULL;

-- 添加索引
ALTER TABLE matches 
ADD INDEX idx_matches_sport_type (sport_type_id);

-- 数据迁移：将现有的tournament数据映射到运动类型
-- 这里假设现有数据都是LOL比赛，映射到LOL运动类型
UPDATE matches 
SET sport_type_id = (SELECT id FROM sport_types WHERE code = 'lol' LIMIT 1)
WHERE sport_type_id IS NULL;