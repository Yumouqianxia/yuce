<template>
  <div class="admin-dashboard">
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

onMounted(() => {
  document.title = '管理后台 | 预测系统'
})
</script>

<style scoped>
.admin-dashboard {
  width: 100%;
  height: 100%;
}
</style> 