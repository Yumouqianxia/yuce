<template>
  <div class="footer">
    <div class="footer-content">
      <div class="footer-links">
        <a href="#" @click.prevent>关于我们</a>
        <a href="#" @click.prevent>联系方式</a>
        <a href="#" @click.prevent>使用条款</a>
        <a href="#" @click.prevent>隐私政策</a>
      </div>
      <div class="footer-copyright">
        <span>© {{ currentYear }} 预测系统 | 赛事预测与积分排行</span>
        <span class="icp-info">粤ICP备2025400738号</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'

const currentYear = computed(() => new Date().getFullYear())
</script>

<style scoped>
.footer {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-light);
  font-size: 13px;
}

.footer-content {
  max-width: 1200px;
  width: 100%;
  text-align: center;
  padding: 0 24px;
}

.footer-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  margin-bottom: 6px;
}

.footer-links a {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}

.footer-links a:hover {
  color: var(--primary-color);
}

.footer-copyright {
  color: var(--text-light);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
}

.icp-info {
  color: var(--text-light);
  font-size: 12px;
}

@media (max-width: 768px) {
  .footer-copyright {
    flex-direction: column;
    gap: 4px;
  }
  
  .icp-info {
    font-size: 11px;
  }
}
</style> 