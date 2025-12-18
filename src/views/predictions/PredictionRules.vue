<template>
  <div class="prediction-list-container">
    <h1 class="page-title">积分规则</h1>

    <!-- 导航菜单 -->
    <div class="prediction-nav">
      <router-link to="/upcoming-matches" class="nav-item">最近比赛</router-link>
      <router-link to="/prediction-history" class="nav-item">我的预测历史</router-link>
      <div class="nav-item active">积分规则</div>
    </div>

    <div class="rules-content">
      <div class="rules-section">
        <h2>预测积分规则说明</h2>
        <p class="rules-intro">
          在电竞预测平台，您可以通过预测比赛结果获取积分。积分将用于排行榜排名，展示您的预测能力。
          以下是积分计算规则的详细说明：
        </p>

        <div class="rule-item">
          <div class="rule-title">完全正确预测</div>
          <div class="rule-points">+5分</div>
          <div class="rule-description">
            预测的获胜队伍和比分都正确。例如，您预测T1以2-1战胜DK，实际结果也是T1以2-1获胜。
          </div>
        </div>

        <div class="rule-item">
          <div class="rule-title">获胜队伍正确</div>
          <div class="rule-points">+3分</div>
          <div class="rule-description">
            预测的获胜队伍正确，但比分不正确。例如，您预测T1以2-0战胜DK，实际结果是T1以2-1获胜。
          </div>
        </div>

        <div class="rule-item">
          <div class="rule-title">比分正确</div>
          <div class="rule-points">+1分</div>
          <div class="rule-description">
            预测的比分正确，但获胜队伍不正确。例如，您预测T1以2-1战胜DK，实际结果是DK以2-1获胜。
          </div>
        </div>

        <div class="rule-item">
          <div class="rule-title">预测错误</div>
          <div class="rule-points">+0分</div>
          <div class="rule-description">
            预测的获胜队伍和比分都不正确。例如，您预测T1以2-0战胜DK，实际结果是DK以2-1获胜。
          </div>
        </div>
      </div>

      <div class="rules-section">
        <h2>预测规则</h2>

        <div class="rule-list">
          <div class="rule-list-item">
            <div class="rule-number">1</div>
            <div class="rule-text">每场比赛可以进行预测，在比赛开始前可以修改预测。</div>
          </div>

          <div class="rule-list-item">
            <div class="rule-number">2</div>
            <div class="rule-text">必须在比赛开始前提交预测，比赛开始后将无法进行预测。</div>
          </div>

          <div class="rule-list-item">
            <div class="rule-number">3</div>
            <div class="rule-text">预测必须包含获胜队伍和具体比分。</div>
          </div>

          <div class="rule-list-item">
            <div class="rule-number">4</div>
            <div class="rule-text">比赛结束后，系统将自动计算积分并更新排行榜。</div>
          </div>

          <div class="rule-list-item">
            <div class="rule-number">5</div>
            <div class="rule-text">如比赛因特殊原因取消或延期，相关预测将被视为无效。</div>
          </div>
        </div>
      </div>

      <div class="rules-note">
        <h3>注意事项</h3>
        <p>
          积分排行榜每日更新，展示用户的累计积分。我们鼓励公平竞争，禁止任何作弊行为。
          如有任何疑问或建议，请联系管理员。祝您预测愉快！
        </p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PredictionRules'
}
</script>

<style scoped>
.prediction-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 30px;
  color: #303133;
  text-align: center;
}

.prediction-nav {
  display: flex;
  justify-content: center;
  margin-bottom: 30px;
  background-color: #fff;
  border-radius: 8px;
  padding: 15px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.nav-item {
  padding: 10px 20px;
  margin: 0 10px;
  color: #606266;
  text-decoration: none;
  font-size: 16px;
  border-radius: 4px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.nav-item:hover {
  color: #409EFF;
  background-color: #ecf5ff;
}

.nav-item.active {
  color: #409EFF;
  font-weight: 500;
  background-color: #ecf5ff;
}

.rules-content {
  background-color: var(--bg-white);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-sm);
  padding: 30px;
  margin-bottom: 40px;
  border: 1px solid transparent;
  transition: all 0.3s;
}

.rules-content:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--primary-light);
}

.rules-section {
  margin-bottom: 40px;
}

.rules-section h2 {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.rules-intro {
  margin-bottom: 25px;
  color: #666;
  line-height: 1.6;
}

.rule-item {
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 15px;
  border-left: 4px solid #3a7bd5;
}

.rule-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 5px;
}

.rule-points {
  font-size: 16px;
  font-weight: 700;
  color: #3a7bd5;
  margin-bottom: 10px;
}

.rule-description {
  color: #666;
  line-height: 1.5;
}

.rule-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.rule-list-item {
  display: flex;
  align-items: flex-start;
}

.rule-number {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background-color: #3a7bd5;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  margin-right: 15px;
  flex-shrink: 0;
}

.rule-text {
  flex: 1;
  padding-top: 5px;
  color: #666;
  line-height: 1.5;
}

.rules-note {
  background-color: #f0f9ff;
  border-radius: 8px;
  padding: 20px;
  border-left: 4px solid #60a5fa;
}

.rules-note h3 {
  font-size: 18px;
  margin-bottom: 10px;
  color: #333;
}

.rules-note p {
  color: #666;
  line-height: 1.6;
  margin: 0;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .prediction-nav {
    flex-direction: column;
    align-items: center;
  }

  .nav-item {
    margin: 5px 0;
    width: 100%;
    text-align: center;
  }

  .page-title {
    font-size: 24px;
    margin-bottom: 20px;
  }

  .rules-content {
    padding: 20px;
  }
}
</style>
