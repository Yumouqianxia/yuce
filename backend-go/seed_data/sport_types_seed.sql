-- 运动类型种子数据
INSERT INTO sport_types (name, code, category, icon, banner, description, is_active, sort_order) VALUES
-- 电子竞技
('英雄联盟', 'lol', 'esports', '/images/sports/lol-icon.png', '/images/sports/lol-banner.jpg', '全球最受欢迎的MOBA游戏，拥有庞大的电竞赛事体系', TRUE, 1),
('王者荣耀', 'wzry', 'esports', '/images/sports/wzry-icon.png', '/images/sports/wzry-banner.jpg', '腾讯旗下热门手机MOBA游戏，国内电竞领军项目', TRUE, 2),
('DOTA2', 'dota2', 'esports', '/images/sports/dota2-icon.png', '/images/sports/dota2-banner.jpg', 'Valve开发的经典MOBA游戏，拥有TI国际邀请赛', TRUE, 3),
('CS2', 'cs2', 'esports', '/images/sports/cs2-icon.png', '/images/sports/cs2-banner.jpg', 'Counter-Strike 2，全球顶级FPS电竞项目', TRUE, 4),
('和平精英', 'hpjy', 'esports', '/images/sports/hpjy-icon.png', '/images/sports/hpjy-banner.jpg', '腾讯旗下战术竞技手游，移动电竞热门项目', TRUE, 5),

-- 传统体育
('足球', 'football', 'traditional', '/images/sports/football-icon.png', '/images/sports/football-banner.jpg', '世界第一运动，拥有世界杯、欧洲杯等顶级赛事', FALSE, 10),
('篮球', 'basketball', 'traditional', '/images/sports/basketball-icon.png', '/images/sports/basketball-banner.jpg', '全球热门运动，NBA、CBA等职业联赛', FALSE, 11),
('羽毛球', 'badminton', 'traditional', '/images/sports/badminton-icon.png', '/images/sports/badminton-banner.jpg', '亚洲热门运动，奥运会正式项目', FALSE, 12);

-- 运动配置种子数据
INSERT INTO sport_configurations (sport_type_id, enable_realtime, enable_chat, enable_voting, enable_prediction, enable_leaderboard, allow_modification, max_modifications, modification_deadline, enable_self_voting, max_votes_per_user, voting_deadline) VALUES
-- 英雄联盟配置 (全功能开启)
(1, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, 3, 30, FALSE, 10, 0),
-- 王者荣耀配置 (全功能开启)
(2, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, 3, 30, FALSE, 10, 0),
-- DOTA2配置 (关闭聊天)
(3, TRUE, FALSE, TRUE, TRUE, TRUE, TRUE, 2, 60, FALSE, 5, 0),
-- CS2配置 (关闭聊天，限制修改)
(4, TRUE, FALSE, TRUE, TRUE, TRUE, TRUE, 1, 15, FALSE, 5, 30),
-- 和平精英配置 (基础配置)
(5, TRUE, FALSE, TRUE, TRUE, TRUE, TRUE, 2, 30, FALSE, 8, 0),
-- 足球配置 (传统体育基础配置)
(6, FALSE, FALSE, TRUE, TRUE, TRUE, FALSE, 0, 0, FALSE, 3, 0),
-- 篮球配置 (传统体育基础配置)
(7, FALSE, FALSE, TRUE, TRUE, TRUE, FALSE, 0, 0, FALSE, 3, 0),
-- 羽毛球配置 (传统体育基础配置)
(8, FALSE, FALSE, TRUE, TRUE, TRUE, FALSE, 0, 0, FALSE, 3, 0);

-- 积分规则种子数据
INSERT INTO scoring_rules (sport_type_id, name, description, is_active, base_points, enable_difficulty, difficulty_multiplier, enable_vote_reward, vote_reward_points, max_vote_reward, enable_time_reward, time_reward_points, time_reward_hours, enable_modify_penalty, modify_penalty_points, max_modify_penalty) VALUES
-- 英雄联盟积分规则 (复杂规则)
(1, 'LOL标准积分规则', '英雄联盟预测积分规则，包含投票奖励和时间奖励', TRUE, 10, TRUE, 1.5, TRUE, 1, 15, TRUE, 5, 24, TRUE, 2, 6),
-- 王者荣耀积分规则 (复杂规则)
(2, '王者荣耀标准积分规则', '王者荣耀预测积分规则，包含投票奖励和时间奖励', TRUE, 10, TRUE, 1.3, TRUE, 1, 12, TRUE, 3, 12, TRUE, 1, 3),
-- DOTA2积分规则 (中等复杂度)
(3, 'DOTA2标准积分规则', 'DOTA2预测积分规则，包含难度系数和时间奖励', TRUE, 15, TRUE, 2.0, FALSE, 0, 0, TRUE, 8, 48, TRUE, 3, 9),
-- CS2积分规则 (简单规则)
(4, 'CS2标准积分规则', 'CS2预测积分规则，基础积分制', TRUE, 8, FALSE, 1.0, TRUE, 1, 8, FALSE, 0, 0, FALSE, 0, 0),
-- 和平精英积分规则 (中等复杂度)
(5, '和平精英标准积分规则', '和平精英预测积分规则，包含投票奖励', TRUE, 12, TRUE, 1.2, TRUE, 1, 10, TRUE, 4, 18, TRUE, 2, 4),
-- 传统体育积分规则 (简单规则)
(6, '足球标准积分规则', '足球预测积分规则，基础积分制', TRUE, 20, FALSE, 1.0, FALSE, 0, 0, FALSE, 0, 0, FALSE, 0, 0),
(7, '篮球标准积分规则', '篮球预测积分规则，基础积分制', TRUE, 15, FALSE, 1.0, FALSE, 0, 0, FALSE, 0, 0, FALSE, 0, 0),
(8, '羽毛球标准积分规则', '羽毛球预测积分规则，基础积分制', TRUE, 12, FALSE, 1.0, FALSE, 0, 0, FALSE, 0, 0, FALSE, 0, 0);