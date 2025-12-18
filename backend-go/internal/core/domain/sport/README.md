# 运动类型管理系统

## 概述

运动类型管理系统为预测平台提供了灵活的运动分类和配置管理功能。支持电子竞技和传统体育两大类别，每种运动类型都可以独立配置功能开关和积分规则。

## 数据模型

### SportType (运动类型)

运动类型是系统的核心分类概念，替代了原有的固定锦标赛概念。

**字段说明:**
- `ID`: 主键
- `Name`: 运动名称 (如: 英雄联盟、王者荣耀、足球)
- `Code`: 运动代码 (如: lol、wzry、football)
- `Category`: 运动类别 (esports/traditional)
- `Icon`: 运动图标URL
- `Banner`: 运动横幅图URL
- `Description`: 运动描述
- `IsActive`: 是否启用
- `SortOrder`: 首页显示顺序

**业务方法:**
- `IsEsports()`: 检查是否为电子竞技
- `IsTraditional()`: 检查是否为传统体育
- `GetDisplayName()`: 获取显示名称
- `HasConfiguration()`: 检查是否有配置

### SportConfiguration (运动配置)

每种运动类型的功能配置，控制该运动的各种功能开关。

**功能开关:**
- `EnableRealtime`: 启用实时通信
- `EnableChat`: 启用聊天功能
- `EnableVoting`: 启用投票功能
- `EnablePrediction`: 启用预测功能
- `EnableLeaderboard`: 启用排行榜

**预测设置:**
- `AllowModification`: 允许修改预测
- `MaxModifications`: 最大修改次数
- `ModificationDeadline`: 修改截止时间(分钟)

**投票设置:**
- `EnableSelfVoting`: 允许给自己投票
- `MaxVotesPerUser`: 每用户最大投票数
- `VotingDeadline`: 投票截止时间(分钟)

**业务方法:**
- `IsFeatureEnabled(feature)`: 检查功能是否启用
- `CanModifyPrediction(count)`: 检查是否可以修改预测
- `CanVote(count, isSelf)`: 检查是否可以投票

### ScoringRule (积分规则)

可配置的积分计算规则，支持多种积分组件的开关控制。

**基础设置:**
- `BasePoints`: 基础积分
- `EnableDifficulty`: 启用难度系数
- `DifficultyMultiplier`: 难度系数

**奖励组件:**
- `EnableVoteReward`: 启用投票奖励
- `VoteRewardPoints`: 每票奖励积分
- `MaxVoteReward`: 最大投票奖励
- `EnableTimeReward`: 启用时间奖励
- `TimeRewardPoints`: 时间奖励积分
- `TimeRewardHours`: 时间奖励小时数

**惩罚组件:**
- `EnableModifyPenalty`: 启用修改惩罚
- `ModifyPenaltyPoints`: 每次修改扣分
- `MaxModifyPenalty`: 最大修改惩罚

## 数据库表结构

### sport_types 表
```sql
CREATE TABLE sport_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(20) NOT NULL UNIQUE,
    category VARCHAR(20) NOT NULL,
    icon VARCHAR(255) DEFAULT '',
    banner VARCHAR(255) DEFAULT '',
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### sport_configurations 表
```sql
CREATE TABLE sport_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sport_type_id BIGINT UNSIGNED NOT NULL UNIQUE,
    enable_realtime BOOLEAN DEFAULT TRUE,
    enable_chat BOOLEAN DEFAULT FALSE,
    enable_voting BOOLEAN DEFAULT TRUE,
    enable_prediction BOOLEAN DEFAULT TRUE,
    enable_leaderboard BOOLEAN DEFAULT TRUE,
    allow_modification BOOLEAN DEFAULT TRUE,
    max_modifications INT DEFAULT 3,
    modification_deadline INT DEFAULT 30,
    enable_self_voting BOOLEAN DEFAULT FALSE,
    max_votes_per_user INT DEFAULT 10,
    voting_deadline INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
);
```

### scoring_rules 表
```sql
CREATE TABLE scoring_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sport_type_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    base_points INT DEFAULT 10,
    enable_difficulty BOOLEAN DEFAULT FALSE,
    difficulty_multiplier DECIMAL(3,2) DEFAULT 1.00,
    enable_vote_reward BOOLEAN DEFAULT FALSE,
    vote_reward_points INT DEFAULT 1,
    max_vote_reward INT DEFAULT 10,
    enable_time_reward BOOLEAN DEFAULT FALSE,
    time_reward_points INT DEFAULT 5,
    time_reward_hours INT DEFAULT 24,
    enable_modify_penalty BOOLEAN DEFAULT FALSE,
    modify_penalty_points INT DEFAULT 2,
    max_modify_penalty INT DEFAULT 6,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
);
```

## 种子数据

系统预置了以下运动类型：

**电子竞技:**
- 英雄联盟 (lol)
- 王者荣耀 (wzry)
- DOTA2 (dota2)
- CS2 (cs2)
- 和平精英 (hpjy)

**传统体育:**
- 足球 (football)
- 篮球 (basketball)
- 羽毛球 (badminton)

每种运动类型都有对应的配置和积分规则设置。

## 与现有系统的集成

### 比赛表更新

现有的 `matches` 表添加了 `sport_type_id` 字段，与运动类型建立关联：

```sql
ALTER TABLE matches 
ADD COLUMN sport_type_id BIGINT UNSIGNED DEFAULT NULL;
```

原有的 `tournament` 字段保留以确保向后兼容性。

### 数据迁移

现有比赛数据会自动迁移到LOL运动类型，确保系统平滑升级。

## 使用示例

```go
// 创建运动类型
sportType := &sport.SportType{
    Name:        "英雄联盟",
    Code:        "lol",
    Category:    sport.SportCategoryEsports,
    Icon:        "/images/sports/lol-icon.png",
    Banner:      "/images/sports/lol-banner.jpg",
    Description: "全球最受欢迎的MOBA游戏",
    IsActive:    true,
    SortOrder:   1,
}

// 检查运动类别
if sportType.IsEsports() {
    // 电子竞技相关逻辑
}

// 创建配置
config := &sport.SportConfiguration{
    SportTypeID:           sportType.ID,
    EnableRealtime:        true,
    EnableChat:           true,
    EnableVoting:         true,
    AllowModification:    true,
    MaxModifications:     3,
    ModificationDeadline: 30,
}

// 检查功能是否启用
if config.IsFeatureEnabled("chat") {
    // 启用聊天功能
}

// 检查是否可以修改预测
if config.CanModifyPrediction(userModificationCount) {
    // 允许修改预测
}
```