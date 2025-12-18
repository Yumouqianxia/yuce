<template>
  <div class="nav-icon">
    <el-icon v-if="icon === 'home'"><HomeFilled /></el-icon>
    <el-icon v-else-if="icon === 'trophy'"><Trophy /></el-icon>
    <el-icon v-else-if="icon === 'list'"><List /></el-icon>
    <el-icon v-else-if="icon === 'cog'"><Setting /></el-icon>
    <el-icon v-else-if="icon === 'chart-bar'"><Trophy /></el-icon>
  </div>
</template>

<script lang="ts" setup>
import { HomeFilled, Trophy, List, Setting } from '@element-plus/icons-vue'

defineProps({
  icon: {
    type: String,
    required: true,
    validator: (value: string) => ['home', 'trophy', 'list', 'cog', 'chart-bar'].includes(value)
  }
})
</script>

<style scoped>
.nav-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 4px;
}

:deep(.el-icon) {
  font-size: 20px; /* 调整图标大小 */
}

@media (max-width: 767px) {
  :deep(.el-icon) {
    font-size: 22px; /* 移动设备上稍大一点 */
  }
}
</style>