<template>
  <div class="admin-section">
    <h2>公告管理</h2>



    <div class="announcement-manager">
      <h3>公告管理</h3>
      <p>在这里可以管理网站的公告信息，首页将显示最新的一条公告。</p>

      <div class="announcement-form">
        <form @submit.prevent="submitAnnouncement" class="form">
          <div class="form-group">
            <label for="title">标题</label>
            <input
              id="title"
              v-model="announcementForm.title"
              type="text"
              class="input-field"
              placeholder="请输入公告标题"
              required
              minlength="2"
              maxlength="50"
            />
            <div v-if="validationErrors.title" class="error-message">{{ validationErrors.title }}</div>
          </div>

          <div class="form-group">
            <label for="content">内容</label>
            <textarea
              id="content"
              v-model="announcementForm.content"
              class="input-field textarea"
              placeholder="请输入公告内容"
              rows="4"
              required
              minlength="5"
              maxlength="500"
            ></textarea>
            <div v-if="validationErrors.content" class="error-message">{{ validationErrors.content }}</div>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '发布中...' : '发布公告' }}
            </button>
          </div>
        </form>
      </div>

      <div class="announcement-list">
        <h4>已发布的公告</h4>
        <div v-if="loading" class="loading-indicator">加载中...</div>
        <div v-else-if="announcements.length === 0" class="empty-state">暂无公告</div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th width="250">标题</th>
              <th>内容</th>
              <th width="180">发布时间</th>
              <th width="100">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="announcement in announcements" :key="announcement.id">
              <td>{{ announcement.title }}</td>
              <td>
                <div class="content-preview">{{ announcement.content }}</div>
              </td>
              <td>{{ formatDate(announcement.createdAt) }}</td>
              <td>
                <button
                  class="btn btn-danger btn-sm"
                  @click="deleteAnnouncement(announcement.id)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

// 加载状态
const loading = ref(false)
const submitting = ref(false)

// 公告表单
const announcementForm = reactive({
  title: '',
  content: ''
})

// 表单验证错误
const validationErrors = reactive({
  title: '',
  content: ''
})

// 定义公告接口
interface Announcement {
  id: number;
  title: string;
  content: string;
  createdAt: string;
  updatedAt?: string;
}

// 公告列表
const announcements = ref<Announcement[]>([])

// 获取公告列表
const fetchAnnouncements = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/announcements', {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    const body = response.data
    if (body && typeof body === 'object') {
      if (body.success === true && body.data) {
        const payload = body.data
        if (Array.isArray(payload.announcements)) {
          announcements.value = payload.announcements
        } else if (Array.isArray(payload)) {
          announcements.value = payload
        } else {
          console.error('无法识别的公告数据格式:', payload)
          announcements.value = []
        }
      } else if (Array.isArray(body)) {
        announcements.value = body
      } else {
        console.error('无法识别的公告响应格式:', body)
        announcements.value = []
      }
    } else {
      console.error('公告响应数据不是对象:', body)
      announcements.value = []
    }
  } catch (error) {
    console.error('获取公告列表失败:', error)
    showMessage('获取公告列表失败', 'error')
  } finally {
    loading.value = false
  }
}

// 验证表单
const validateForm = () => {
  let isValid = true

  // 重置验证错误
  validationErrors.title = ''
  validationErrors.content = ''

  // 验证标题
  if (!announcementForm.title) {
    validationErrors.title = '请输入公告标题'
    isValid = false
  } else if (announcementForm.title.length < 2) {
    validationErrors.title = '标题长度不能少于2个字符'
    isValid = false
  } else if (announcementForm.title.length > 50) {
    validationErrors.title = '标题长度不能超过50个字符'
    isValid = false
  }

  // 验证内容
  if (!announcementForm.content) {
    validationErrors.content = '请输入公告内容'
    isValid = false
  } else if (announcementForm.content.length < 5) {
    validationErrors.content = '内容长度不能少于5个字符'
    isValid = false
  } else if (announcementForm.content.length > 500) {
    validationErrors.content = '内容长度不能超过500个字符'
    isValid = false
  }

  return isValid
}

// 提交公告
const submitAnnouncement = async () => {
  if (!validateForm()) return

  submitting.value = true
  try {
    await axios.post('/api/announcements', announcementForm, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    showMessage('公告发布成功', 'success')
    // 重置表单
    announcementForm.title = ''
    announcementForm.content = ''
    // 重新获取公告列表
    fetchAnnouncements()
  } catch (error) {
    console.error('发布公告失败:', error)
    showMessage('发布公告失败', 'error')
  } finally {
    submitting.value = false
  }
}

// 删除公告
const deleteAnnouncement = async (id: number) => {
  if (!confirm('确定要删除这条公告吗？删除后无法恢复。')) {
    return
  }

  loading.value = true
  try {
    await axios.delete(`/api/announcements/${id}`, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    showMessage('公告删除成功', 'success')
    // 重新获取公告列表
    fetchAnnouncements()
  } catch (error) {
    console.error('删除公告失败:', error)
    showMessage('删除公告失败', 'error')
  } finally {
    loading.value = false
  }
}

// 显示消息
const showMessage = (message: string, type = 'info') => {
  // 创建消息元素
  const messageElement = document.createElement('div')
  messageElement.className = `message message-${type}`
  messageElement.textContent = message

  // 添加到文档
  document.body.appendChild(messageElement)

  // 显示消息
  setTimeout(() => {
    messageElement.classList.add('show')
  }, 10)

  // 定时移除
  setTimeout(() => {
    messageElement.classList.remove('show')
    setTimeout(() => {
      document.body.removeChild(messageElement)
    }, 300)
  }, 3000)
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return ''

  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(() => {
  document.title = '公告管理 | 预测系统'
  fetchAnnouncements()
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

.dashboard-stats {
  display: flex;
  justify-content: space-between;
  margin-bottom: 30px;
}

.stat-card {
  flex: 1;
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  margin: 0 10px;
  text-align: center;
}

.stat-card:first-child {
  margin-left: 0;
}

.stat-card:last-child {
  margin-right: 0;
}

.stat-card h3 {
  margin-top: 0;
  color: #606266;
  font-size: 16px;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  margin: 10px 0;
  color: #409EFF;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

/* 公告管理样式 */
.announcement-manager {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.announcement-manager h3 {
  margin-top: 0;
  margin-bottom: 15px;
  font-size: 18px;
  color: #303133;
}

.announcement-manager p {
  color: #606266;
  margin-bottom: 20px;
}

/* 表单样式 */
.announcement-form {
  margin-top: 20px;
  margin-bottom: 30px;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.form {
  width: 100%;
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

.textarea {
  min-height: 100px;
  resize: vertical;
}

.error-message {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 5px;
}

.form-actions {
  margin-top: 20px;
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

.btn-danger {
  background-color: #f56c6c;
  color: white;
  border-color: #f56c6c;
}

.btn-danger:hover {
  background-color: #f78989;
  border-color: #f78989;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

/* 公告列表样式 */
.announcement-list {
  margin-top: 30px;
}

.announcement-list h4 {
  margin-top: 0;
  margin-bottom: 15px;
  font-size: 16px;
  color: #606266;
}

.loading-indicator {
  text-align: center;
  padding: 20px;
  color: #909399;
}

.empty-state {
  text-align: center;
  padding: 30px;
  color: #909399;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  border-spacing: 0;
  margin-bottom: 20px;
}

.data-table th,
.data-table td {
  padding: 12px 8px;
  text-align: left;
  border-bottom: 1px solid #ebeef5;
}

.data-table th {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 500;
  white-space: nowrap;
}

.data-table tr:hover td {
  background-color: #f5f7fa;
}

.content-preview {
  max-height: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 响应式样式 */
@media (max-width: 768px) {
  .dashboard-stats {
    flex-direction: column;
  }

  .stat-card {
    margin: 0 0 15px 0;
  }

  .data-table th,
  .data-table td {
    padding: 8px 4px;
    font-size: 13px;
  }
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