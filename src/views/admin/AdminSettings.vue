<template>
  <div class="admin-section">
    <h2>系统设置</h2>
    <form @submit.prevent="saveSettings" class="settings-form">
      <div class="form-group">
        <label for="siteName">系统名称</label>
        <input
          id="siteName"
          v-model="systemSettings.siteName"
          type="text"
          class="input-field"
          placeholder="请输入系统名称"
        />
      </div>

      <div class="form-group">
        <label for="allowRegistration">注册是否开放</label>
        <div class="toggle-switch">
          <input
            id="allowRegistration"
            type="checkbox"
            v-model="systemSettings.allowRegistration"
            class="toggle-input"
          />
          <label for="allowRegistration" class="toggle-label"></label>
          <span class="toggle-status">{{ systemSettings.allowRegistration ? '已开放' : '已关闭' }}</span>
        </div>
      </div>

      <div class="form-group">
        <label for="enableLeaderboard">是否开放排行榜</label>
        <div class="toggle-switch">
          <input
            id="enableLeaderboard"
            type="checkbox"
            v-model="systemSettings.enableLeaderboard"
            class="toggle-input"
          />
          <label for="enableLeaderboard" class="toggle-label"></label>
          <span class="toggle-status">{{ systemSettings.enableLeaderboard ? '已开放' : '已关闭' }}</span>
        </div>
      </div>

      <div class="form-group">
        <label for="predictionDeadlineHours">预测截止时间</label>
        <div class="number-input-container">
          <button
            type="button"
            class="number-btn"
            @click="decrementHours"
            :disabled="systemSettings.predictionDeadlineHours <= 0"
          >-</button>
          <input
            id="predictionDeadlineHours"
            v-model.number="systemSettings.predictionDeadlineHours"
            type="number"
            class="number-input"
            min="0"
            max="48"
            step="1"
          />
          <button
            type="button"
            class="number-btn"
            @click="incrementHours"
            :disabled="systemSettings.predictionDeadlineHours >= 48"
          >+</button>
        </div>
        <span class="form-hint">比赛开始前几小时停止预测</span>
      </div>

      <div class="form-actions">
        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSystemSettings, updateSystemSettings } from '@/api/admin'

const saving = ref(false)

const systemSettings = reactive({
  siteName: '预测系统',
  allowRegistration: true,
  enableLeaderboard: true,
  predictionDeadlineHours: 1
})

// 增加预测截止时间
const incrementHours = () => {
  if (systemSettings.predictionDeadlineHours < 48) {
    systemSettings.predictionDeadlineHours++
  }
}

// 减少预测截止时间
const decrementHours = () => {
  if (systemSettings.predictionDeadlineHours > 0) {
    systemSettings.predictionDeadlineHours--
  }
}

// 获取系统设置
const fetchSettings = async () => {
  try {
    const data = await getSystemSettings()
    if (data) {
      Object.assign(systemSettings, data)
    }
  } catch (error) {
    ElMessage.error('获取系统设置失败')
  }
}

// 保存系统设置
const saveSettings = async () => {
  saving.value = true
  try {
    // 打印要发送的数据
    console.log('发送的设置数据:', systemSettings)

    // 确保数字类型正确
    const settingsToSend = {
      ...systemSettings,
      predictionDeadlineHours: Number(systemSettings.predictionDeadlineHours)
    }

    await updateSystemSettings(settingsToSend)
    ElMessage.success('系统设置保存成功')
  } catch (error) {
    ElMessage.error('保存系统设置失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  document.title = '系统设置 | 预测系统'
  fetchSettings()
})
</script>

<style scoped>
.admin-section {
  width: 100%;
}

.admin-section h2 {
  margin-top: 0;
  margin-bottom: 20px;
  font-size: 18px;
  color: #303133;
}

/* 表单样式 */
.settings-form {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #303133;
}

.input-field {
  width: 100%;
  max-width: 400px;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  color: #606266;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input-field:focus {
  outline: none;
  border-color: #409EFF;
}

/* 开关样式 */
.toggle-switch {
  display: flex;
  align-items: center;
}

.toggle-input {
  display: none;
}

.toggle-label {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
  background-color: #dcdfe6;
  border-radius: 10px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.toggle-label:after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background-color: white;
  border-radius: 50%;
  transition: left 0.3s;
}

.toggle-input:checked + .toggle-label {
  background-color: #409EFF;
}

.toggle-input:checked + .toggle-label:after {
  left: 22px;
}

.toggle-status {
  margin-left: 10px;
  font-size: 14px;
  color: #606266;
}

/* 数字输入框样式 */
.number-input-container {
  display: flex;
  align-items: center;
  max-width: 200px;
}

.number-input {
  width: 60px;
  text-align: center;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 0;
  font-size: 14px;
  color: #606266;
  -moz-appearance: textfield; /* Firefox */
}

.number-input::-webkit-outer-spin-button,
.number-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.number-btn {
  width: 30px;
  height: 34px;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  color: #606266;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.number-btn:first-child {
  border-radius: 4px 0 0 4px;
  border-right: none;
}

.number-btn:last-child {
  border-radius: 0 4px 4px 0;
  border-left: none;
}

.number-btn:hover {
  background-color: #e9ecef;
}

.number-btn:disabled {
  background-color: #f5f7fa;
  color: #c0c4cc;
  cursor: not-allowed;
}

.form-hint {
  display: block;
  margin-top: 5px;
  margin-left: 0;
  color: #909399;
  font-size: 14px;
}

/* 按钮样式 */
.form-actions {
  margin-top: 30px;
}

.btn {
  display: inline-block;
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 500;
  text-align: center;
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary {
  background-color: #409EFF;
  color: white;
  border-color: #409EFF;
}

.btn-primary:hover {
  background-color: #66b1ff;
  border-color: #66b1ff;
}

.btn-primary:disabled {
  background-color: #a0cfff;
  border-color: #a0cfff;
  cursor: not-allowed;
}

/* 消息提示样式 */
:global(.message) {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%) translateY(-100px);
  padding: 10px 20px;
  border-radius: 4px;
  color: white;
  font-size: 14px;
  z-index: 9999;
  transition: transform 0.3s ease;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

:global(.message.show) {
  transform: translateX(-50%) translateY(0);
}

:global(.message-success) {
  background-color: #67c23a;
}

:global(.message-error) {
  background-color: #f56c6c;
}

:global(.message-info) {
  background-color: #909399;
}

:global(.message-warning) {
  background-color: #e6a23c;
}
</style>