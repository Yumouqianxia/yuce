-- 删除运动类型关联
ALTER TABLE matches DROP FOREIGN KEY fk_matches_sport_type;
ALTER TABLE matches DROP INDEX idx_matches_sport_type;
ALTER TABLE matches DROP COLUMN sport_type_id;