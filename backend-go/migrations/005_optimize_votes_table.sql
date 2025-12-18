-- 优化投票表索引和约束
-- 确保投票表有正确的索引以提高查询性能

-- 添加复合索引以优化常见查询
CREATE INDEX IF NOT EXISTS idx_votes_prediction_created ON votes(prediction_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_votes_user_created ON votes(user_id, created_at DESC);

-- 确保唯一约束存在（防重复投票）
-- 这个约束应该已经在 004 迁移中创建，但为了确保一致性再次检查
ALTER TABLE votes ADD CONSTRAINT uk_user_prediction_vote UNIQUE (user_id, prediction_id);

-- 添加外键约束检查（如果不存在）
-- 这些约束应该已经在 004 迁移中创建，但为了确保一致性再次检查
ALTER TABLE votes 
ADD CONSTRAINT fk_votes_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE votes 
ADD CONSTRAINT fk_votes_prediction_id 
FOREIGN KEY (prediction_id) REFERENCES predictions(id) ON DELETE CASCADE;

-- 添加投票数统计触发器以保持数据一致性
-- 当投票被插入时，自动增加预测的投票数
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS tr_votes_insert_update_count
AFTER INSERT ON votes
FOR EACH ROW
BEGIN
    UPDATE predictions 
    SET vote_count = vote_count + 1 
    WHERE id = NEW.prediction_id;
END$$

-- 当投票被删除时，自动减少预测的投票数
CREATE TRIGGER IF NOT EXISTS tr_votes_delete_update_count
AFTER DELETE ON votes
FOR EACH ROW
BEGIN
    UPDATE predictions 
    SET vote_count = GREATEST(vote_count - 1, 0) 
    WHERE id = OLD.prediction_id;
END$$
DELIMITER ;