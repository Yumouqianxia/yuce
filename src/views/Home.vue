<template>
  <div class="home-container">
    <!-- 欢迎横幅 -->
    <section class="welcome-banner">
      <div class="banner-content">
        <h1>欢迎来到电竞预测平台</h1>
        <p>预测比赛结果，赢取积分，登上排行榜！</p>
        <div class="banner-actions">
          <router-link :to="{ name: 'PredictionUpcomingMatches' }" class="primary-button">立即预测</router-link>
          <router-link to="/prediction-rules" class="secondary-button">了解规则</router-link>
        </div>
      </div>
    </section>

    <!-- 公告模块 -->
    <section class="announcement-section" v-if="latestAnnouncement">
      <div class="section-header">
        <h2>系统公告</h2>
      </div>
      <div class="announcement-content">
        <div class="announcement-item">
          <div class="announcement-icon">📢</div>
          <div class="announcement-text">
            <h3>{{ latestAnnouncement.title }}</h3>
            <p>{{ latestAnnouncement.content }}</p>
            <span class="announcement-date">{{ formatDate(latestAnnouncement.createdAt) }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 无公告时的占位内容 -->
    <section class="announcement-section" v-else>
      <div class="section-header">
        <h2>系统公告</h2>
      </div>
      <div class="announcement-content">
        <div class="announcement-item empty-announcement">
          <p>暂无公告</p>
        </div>
      </div>
    </section>

    <!-- 快速导航 -->
    <section class="quick-nav-section">
      <div class="section-header">
        <h2>快速导航</h2>
      </div>
      <div class="nav-grid">
        <router-link :to="{ name: 'PredictionUpcomingMatches' }" class="nav-item">
          <div class="nav-icon">🎮</div>
          <div class="nav-label">前往预测</div>
        </router-link>

        <router-link to="/matches" class="nav-item">
          <div class="nav-icon">📅</div>
          <div class="nav-label">比赛日程</div>
        </router-link>

        <router-link to="/prediction-history" class="nav-item">
          <div class="nav-icon">📊</div>
          <div class="nav-label">预测历史</div>
        </router-link>

        <router-link to="/leaderboard" class="nav-item">
          <div class="nav-icon">🏆</div>
          <div class="nav-label">积分排行</div>
        </router-link>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const latestAnnouncement = ref(null)

// 获取最新公告
const fetchLatestAnnouncement = async () => {
  try {
    const response = await axios.get('/api/announcements/latest')
    if (response.data && (response.data.data || response.data)) {
      latestAnnouncement.value = response.data.data || response.data
    }
  } catch (error) {
    console.error('获取公告失败:', error)
  }
}

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''

  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

onMounted(() => {
  fetchLatestAnnouncement()
})
</script>

<style scoped>
.home-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

/* 欢迎横幅 */
.welcome-banner {
  background: linear-gradient(135deg, #3a7bd5, #00d2ff);
  border-radius: 12px;
  padding: 40px;
  margin: 20px 0 30px;
  color: white;
  text-align: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.banner-content h1 {
  font-size: 32px;
  margin-bottom: 10px;
  font-weight: 700;
}

.banner-content p {
  font-size: 18px;
  margin-bottom: 25px;
  opacity: 0.9;
}

.banner-actions {
  display: flex;
  justify-content: center;
  gap: 15px;
}

.primary-button, .secondary-button {
  display: inline-block;
  padding: 12px 24px;
  border-radius: 6px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s ease;
}

.primary-button {
  background-color: white;
  color: #3a7bd5;
}

.primary-button:hover {
  background-color: #f0f0f0;
  transform: translateY(-2px);
}

.secondary-button {
  background-color: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.4);
}

.secondary-button:hover {
  background-color: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

/* 公告模块 */
.announcement-section {
  background-color: white;
  border-radius: 12px;
  padding: 25px;
  margin-bottom: 30px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.announcement-section h2 {
  font-size: 22px;
  font-weight: 600;
  color: #333;
  margin: 0 0 20px 0;
}

.announcement-item {
  display: flex;
  align-items: flex-start;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.announcement-icon {
  font-size: 24px;
  margin-right: 15px;
  color: #3a7bd5;
}

.announcement-text h3 {
  font-size: 18px;
  margin: 0 0 10px 0;
  color: #333;
}

.announcement-text p {
  margin: 0 0 10px 0;
  color: #666;
  line-height: 1.5;
}

.announcement-date {
  display: block;
  font-size: 14px;
  color: #999;
  text-align: right;
}

.empty-announcement {
  text-align: center;
  color: #999;
  padding: 20px;
  background-color: #f9f9f9;
}

/* 快速导航 */
.quick-nav-section {
  background-color: white;
  border-radius: 12px;
  padding: 25px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.quick-nav-section h2 {
  font-size: 22px;
  font-weight: 600;
  color: #333;
  margin: 0 0 20px 0;
}

.nav-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 20px;
  text-decoration: none;
  color: #333;
  transition: transform 0.2s, background-color 0.2s;
}

.nav-item:hover {
  background-color: #f0f0f0;
  transform: translateY(-3px);
}

.nav-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.nav-label {
  font-weight: 500;
  font-size: 16px;
}

@media (max-width: 768px) {
  .welcome-banner {
    padding: 30px;
  }

  .banner-content h1 {
    font-size: 24px;
  }

  .nav-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>