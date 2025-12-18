<template>
  <div class="profile-edit-container">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="profile-edit-form-container">
      <div class="header">
        <h2 class="title">编辑个人资料</h2>
        <button class="back-button" @click="goBack">返回个人资料</button>
      </div>

      <form @submit.prevent="updateUserProfile" class="form">
        <div class="avatar-section">
          <div class="avatar">
            <img v-if="userInfo.avatar" :src="userInfo.avatar" alt="用户头像" @error="handleAvatarError">
            <div v-else class="avatar-placeholder">
              {{ getInitial(userInfo.nickname || userInfo.username) }}
            </div>
          </div>
          <div class="avatar-upload">
            <label for="avatar-input" class="upload-button">更换头像</label>
            <input
              id="avatar-input"
              type="file"
              accept="image/*"
              @change="handleAvatarChange"
              style="display: none;"
            >
            <p class="upload-hint">支持 JPG、PNG、GIF 格式，文件大小不超过 2MB</p>
          </div>
        </div>

        <div class="form-item">
          <label>用户名</label>
          <input
            v-model="userForm.username"
            type="text"
            disabled
          />
        </div>

        <div class="form-item">
          <label>昵称</label>
          <input
            v-model="userForm.nickname"
            type="text"
            placeholder="请输入昵称"
          />
          <div v-if="errors.nickname" class="error-message">{{ errors.nickname }}</div>
        </div>

        <div class="form-item">
          <label>电子邮箱</label>
          <input
            v-model="userForm.email"
            type="email"
            placeholder="请输入电子邮箱"
          />
          <div v-if="errors.email" class="error-message">{{ errors.email }}</div>
        </div>

        <div class="password-section-toggle">
          <button
            type="button"
            class="toggle-button"
            @click="showPasswordSection = !showPasswordSection"
          >
            {{ showPasswordSection ? '隐藏密码设置' : '修改密码' }}
          </button>
        </div>

        <div v-if="showPasswordSection" class="password-section">
          <h3 class="section-title">修改密码</h3>

          <div class="form-item">
            <label>当前密码</label>
            <input
              v-model="passwordForm.currentPassword"
              type="password"
              placeholder="请输入当前密码"
            />
            <div v-if="errors.currentPassword" class="error-message">{{ errors.currentPassword }}</div>
          </div>

          <div class="form-item">
            <label>新密码</label>
            <input
              v-model="passwordForm.newPassword"
              type="password"
              placeholder="请输入新密码"
            />
            <div v-if="errors.newPassword" class="error-message">{{ errors.newPassword }}</div>
          </div>

          <div class="form-item">
            <label>确认新密码</label>
            <input
              v-model="passwordForm.confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
            />
            <div v-if="errors.confirmPassword" class="error-message">{{ errors.confirmPassword }}</div>
          </div>

          <div class="password-actions">
            <button
              type="button"
              class="password-button"
              @click="updatePassword"
              :disabled="changingPassword"
            >
              {{ changingPassword ? '更新中...' : '更新密码' }}
            </button>
          </div>
        </div>

        <div class="form-actions">
          <button
            type="submit"
            class="submit-button"
            :disabled="submitting"
          >
            {{ submitting ? '保存中...' : '保存个人资料' }}
          </button>
          <button
            type="button"
            class="cancel-button"
            @click="goBack"
          >
            取消
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getUserProfile, updateProfile } from '@/api/user'
import type { UserProfile } from '@/types/user'
import { getFullAvatarUrl, addTimestamp } from '@/utils/url'
import axios from 'axios'

const userStore = useUserStore()
const router = useRouter()

// 状态
const loading = ref(true)
const submitting = ref(false)
const changingPassword = ref(false)
const showPasswordSection = ref(false)
// 不需要使用defaultAvatar

// 用户信息
const userInfo = ref<UserProfile>({
  username: '',
  nickname: '',
  email: ''
})

// 表单数据
const userForm = reactive({
  username: '',
  nickname: '',
  email: ''
})

// 错误信息
const errors = reactive({
  nickname: '',
  email: '',
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 获取用户信息
const fetchUserProfile = async () => {
  loading.value = true
  try {
    // 获取用户信息
    const userData = await getUserProfile()
    console.log('获取到的用户数据:', userData)

    // 确保头像 URL 是正确的
    if (userData.avatar) {
      // 使用工具函数获取完整的头像 URL
      userData.avatar = getFullAvatarUrl(userData.avatar)
      console.log('头像 URL 已修正为绝对路径:', userData.avatar)
    }

    // 使用返回的用户数据
    userInfo.value = userData

    // 填充表单
    userForm.username = userData.username || ''
    userForm.nickname = userData.nickname || ''
    userForm.email = userData.email || ''
  } catch (error) {
    console.error('获取用户信息失败:', error)
    showMessage('获取用户信息失败', 'error')
  } finally {
    loading.value = false
  }
}

// 验证表单
const validateForm = () => {
  let isValid = true

  // 重置错误信息
  errors.nickname = ''
  errors.email = ''

  // 验证昵称
  if (userForm.nickname && (userForm.nickname.length < 2 || userForm.nickname.length > 20)) {
    errors.nickname = '昵称长度应为2-20个字符'
    isValid = false
  }

  // 验证邮箱
  if (!userForm.email) {
    errors.email = '请输入邮箱'
    isValid = false
  } else if (!/^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/.test(userForm.email)) {
    errors.email = '邮箱格式不正确'
    isValid = false
  }

  return isValid
}

// 验证密码表单
const validatePasswordForm = () => {
  let isValid = true

  // 重置错误信息
  errors.currentPassword = ''
  errors.newPassword = ''
  errors.confirmPassword = ''

  // 验证当前密码
  if (!passwordForm.currentPassword) {
    errors.currentPassword = '请输入当前密码'
    isValid = false
  } else if (passwordForm.currentPassword.length < 6 || passwordForm.currentPassword.length > 20) {
    errors.currentPassword = '密码长度应为6-20个字符'
    isValid = false
  }

  // 验证新密码
  if (!passwordForm.newPassword) {
    errors.newPassword = '请输入新密码'
    isValid = false
  } else if (passwordForm.newPassword.length < 6 || passwordForm.newPassword.length > 20) {
    errors.newPassword = '密码长度应为6-20个字符'
    isValid = false
  }

  // 验证确认密码
  if (!passwordForm.confirmPassword) {
    errors.confirmPassword = '请再次输入新密码'
    isValid = false
  } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
    errors.confirmPassword = '两次输入的密码不一致'
    isValid = false
  }

  return isValid
}

// 更新用户资料
const updateUserProfile = async () => {
  if (!validateForm()) return

  submitting.value = true
  try {
    await updateProfile({
      nickname: userForm.nickname,
      email: userForm.email
    })

    // 更新用户信息
    userInfo.value.nickname = userForm.nickname
    userInfo.value.email = userForm.email

    // 更新全局状态（保持头像等其他字段）
    userStore.updateUserInfo({
      nickname: userForm.nickname,
      email: userForm.email,
      avatar: userInfo.value.avatar
    })

    showMessage('个人资料更新成功', 'success')

    // 保存成功后跳转回个人资料页
    router.push('/profile')
  } catch (error) {
    if (error instanceof Error) {
      showMessage(error.message || '更新个人资料失败', 'error')
    } else {
      showMessage('更新个人资料失败', 'error')
    }
  } finally {
    submitting.value = false
  }
}

// 处理头像上传 - 完全重写以避免循环请求问题
const handleAvatarChange = async (event: Event) => {
  // 阻止默认事件
  event.preventDefault()
  event.stopPropagation()

  // 获取文件输入元素
  const input = event.target as HTMLInputElement

  // 如果没有文件，直接返回
  if (!input.files || input.files.length === 0) {
    input.value = ''
    return
  }

  // 获取文件
  const file = input.files[0]

  // 立即清空文件输入，防止重复提交
  input.value = ''

  // 验证文件类型
  const isValidType = ['image/jpeg', 'image/png', 'image/jpg', 'image/gif'].includes(file.type)
  if (!isValidType) {
    showMessage('头像只能是 JPG/PNG/GIF 格式!', 'error')
    return
  }

  // 验证文件大小
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    showMessage('头像大小不能超过 2MB!', 'error')
    return
  }

  try {
    // 显示上传中提示
    showMessage('头像上传中...', 'info')

    // 创建表单数据
    const formData = new FormData()
    formData.append('file', file)

    // 禁用所有表单元素，防止重复提交
    const formElements = document.querySelectorAll('button, input')
    formElements.forEach((el) => {
      if (el instanceof HTMLElement) {
        el.setAttribute('disabled', 'disabled')
      }
    })

    // 使用原生 fetch API 而不是 axios，减少中间件干扰
    const response = await fetch('/api/uploads/avatar', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${userStore.token}`
      },
      body: formData
    })

    // 重新启用表单元素
    formElements.forEach((el) => {
      if (el instanceof HTMLElement) {
        el.removeAttribute('disabled')
      }
    })

    // 如果响应不成功，抛出错误
    if (!response.ok) {
      throw new Error(`上传失败: ${response.status} ${response.statusText}`)
    }

    // 解析响应
    const result = await response.json()

    // 更新头像 URL
    const avatarUrl = result.avatarUrl
    if (!avatarUrl) {
      throw new Error('服务器没有返回有效的头像 URL')
    }

    console.log('服务器返回的头像 URL:', avatarUrl)
    console.log('服务器返回的完整响应:', result)

    // 直接使用文件名构建 API URL
    const filename = result.filename || avatarUrl.split('/').pop()
    console.log('提取的文件名:', filename)

    // 直接构建 API URL，并附加时间戳避免缓存
    const fullAvatarUrl = `/api/uploads/avatar/${filename}`
    const cacheBustUrl = addTimestamp(fullAvatarUrl)
    console.log('生成的完整头像 URL:', cacheBustUrl)

    // 调用资料更新接口，持久化头像
    try {
      await updateProfile({ avatar: fullAvatarUrl })
    } catch (e) {
      console.warn('头像已上传但更新资料接口调用失败，继续使用本地URL', e)
    }

    // 更新组件内的头像 URL
    userInfo.value.avatar = cacheBustUrl

    // 更新全局状态
    userStore.updateUserInfo({
      avatar: cacheBustUrl
    })

    // 显示成功消息
    showMessage('头像更新成功', 'success')

    // 手动更新头像显示，不使用动态加载
    const avatarImg = document.querySelector('.avatar img')
    if (avatarImg) {
      // 清除之前的重试计数
      avatarImg.removeAttribute('data-retry-count')

      ;(avatarImg as HTMLImageElement).src = cacheBustUrl
    }
  } catch (error) {
    // 重新启用表单元素
    const formElements = document.querySelectorAll('button, input')
    formElements.forEach((el) => {
      if (el instanceof HTMLElement) {
        el.removeAttribute('disabled')
      }
    })

    // 显示错误消息
    if (error instanceof Error) {
      showMessage(error.message || '头像上传失败', 'error')
    } else {
      showMessage('头像上传失败', 'error')
    }

    console.error('头像上传错误:', error)
  }
}

// 显示消息
const showMessage = (message: string, type: 'success' | 'error' | 'info' | 'warning') => {
  const messageElement = document.createElement('div')
  messageElement.className = `message ${type}`
  messageElement.textContent = message

  document.body.appendChild(messageElement)

  // 3秒后移除消息
  setTimeout(() => {
    messageElement.classList.add('fade-out')
    setTimeout(() => {
      document.body.removeChild(messageElement)
    }, 300)
  }, 3000)
}

// 获取昵称或用户名的首字母
const getInitial = (name: string): string => {
  if (!name) return '?'

  // 如果是中文，返回第一个字
  if (/[\u4e00-\u9fa5]/.test(name.charAt(0))) {
    return name.charAt(0)
  }

  // 如果是英文，返回首字母大写
  return name.charAt(0).toUpperCase()
}

// 处理头像加载错误 - 增强版本
const handleAvatarError = (event: Event) => {
  // 阻止默认事件
  event.preventDefault()
  event.stopPropagation()

  const img = event.target as HTMLImageElement
  console.error('头像加载失败:', img.src)

  // 打印当前用户信息中的头像 URL
  console.log('当前用户信息中的头像 URL:', userInfo.value.avatar)
  console.log('全局状态中的头像 URL:', userStore.user?.avatar)

  // 获取重试次数
  const retryCount = parseInt(img.getAttribute('data-retry-count') || '0')

  // 如果已经重试超过2次，使用默认头像
  if (retryCount >= 2) {
    console.log('头像加载失败次数过多，使用默认头像')
    userInfo.value.avatar = ''
    return
  }

  // 增加重试计数
  img.setAttribute('data-retry-count', (retryCount + 1).toString())

  // 提取文件名
  let filename = ''
  if (img.src.includes('?')) {
    // 如果 URL 包含查询参数，删除它们
    const baseUrl = img.src.split('?')[0]
    filename = baseUrl.split('/').pop() || ''
  } else {
    filename = img.src.split('/').pop() || ''
  }

  // 如果文件名无效，使用默认头像
  if (!filename || filename === 'null' || filename === 'undefined') {
    console.log('文件名无效，使用默认头像')
    userInfo.value.avatar = ''
    return
  }

  // 直接构建API URL，而不是使用工具函数
  const newUrl = `/api/uploads/avatar/${filename}`
  console.log('直接构建的新头像 URL:', newUrl)

  // 添加时间戳避免缓存
  const finalUrl = `${newUrl}?t=${Date.now()}`
  console.log('尝试新的头像 URL:', finalUrl)

  // 更新头像 URL
  img.src = finalUrl
  userInfo.value.avatar = newUrl  // 存储不带时间戳的URL

  // 更新全局状态
  userStore.updateUserInfo({
    avatar: newUrl  // 存储不带时间戳的URL
  })
}

// 更新密码
const updatePassword = async () => {
  if (!validatePasswordForm()) return

  changingPassword.value = true
  try {
    // 打印调试信息
    console.log('尝试更新密码，参数:', {
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword,
      newPasswordConfirm: passwordForm.confirmPassword
    })

    // 调用统一的修改密码接口
    const response = await axios.post(`/api/auth/change-password`, {
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword,
      newPasswordConfirm: passwordForm.confirmPassword
    }, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    console.log('密码更新响应:', response.data)

    // 清空表单
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''

    // 隐藏密码设置区域
    showPasswordSection.value = false

    showMessage('密码更新成功', 'success')
  } catch (error) {
    console.error('更新密码失败:', error)

    if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response) {
      console.error('错误响应数据:', (error.response as any).data)

      // 打印更详细的错误信息
      if ((error.response as any).data && (error.response as any).data.details) {
        console.error('验证错误详情:', (error.response as any).data.details)
      }

      if ((error.response as any).data && (error.response as any).data.message) {
        showMessage((error.response as any).data.message, 'error')
      } else if ((error.response as any).data && (error.response as any).data.error) {
        showMessage((error.response as any).data.error, 'error')
      } else {
        showMessage('更新密码失败', 'error')
      }
    } else if (error instanceof Error) {
      showMessage(error.message || '更新密码失败', 'error')
    } else {
      showMessage('更新密码失败', 'error')
    }
  } finally {
    changingPassword.value = false
  }
}

// 返回个人资料页
const goBack = () => {
  router.push('/profile')
}

// 初始化
// 清理可能导致问题的缓存数据
const cleanupCache = () => {
  try {
    // 清除与头像相关的缓存
    const userData = localStorage.getItem('user')
    if (userData) {
      const user = JSON.parse(userData)
      if (user && user.avatar) {
        // 删除头像属性，强制重新加载
        delete user.avatar
        localStorage.setItem('user', JSON.stringify(user))
        console.log('已清除用户头像缓存')
      }
    }

    // 清除所有头像相关的本地存储
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && (key.includes('avatar') || key.includes('image'))) {
        localStorage.removeItem(key)
        console.log(`已清除缓存项: ${key}`)
      }
    }
  } catch (error) {
    console.error('清除缓存失败:', error)
  }
}

// 修复头像 URL的函数
const fixAvatarUrl = (url: string): string => {
  if (!url) return '';

  // 提取文件名
  const filename = url.split('/').pop() || '';

  // 直接构建 API URL
  return `/api/uploads/avatar/${filename}`;
}

onMounted(() => {
  // 清理缓存
  cleanupCache()

  // 获取用户信息
  fetchUserProfile()

  // 如果已经有头像在全局状态中，确保它使用正确的URL
  const currentUser = userStore.user
  if (currentUser && currentUser.avatar) {
    const fixedUrl = fixAvatarUrl(currentUser.avatar)
    if (fixedUrl !== currentUser.avatar) {
      userStore.updateUserInfo({ avatar: fixedUrl })
      console.log('已修正全局状态中的头像 URL:', fixedUrl)
    }
  }
})
</script>

<style scoped>
.profile-edit-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 20px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #4DA1FF;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 10px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.profile-edit-form-container {
  width: 100%;
  max-width: 550px;
  padding: 40px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.title {
  margin: 0;
  color: #333;
  font-size: 24px;
  font-weight: 500;
}

.back-button {
  padding: 8px 16px;
  background-color: transparent;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.back-button:hover {
  border-color: #c6e2ff;
  color: #4DA1FF;
  background-color: #ecf5ff;
}

.form {
  width: 100%;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  overflow: hidden;
  margin-bottom: 15px;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #4DA1FF;
  color: white;
  font-size: 36px;
  font-weight: bold;
}

.avatar-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.upload-button {
  padding: 8px 16px;
  background-color: #4DA1FF;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-bottom: 8px;
}

.upload-button:hover {
  background-color: #3A90F8;
}

.upload-hint {
  font-size: 12px;
  color: #909399;
  margin: 5px 0 0 0;
}

.form-item {
  margin-bottom: 25px;
  width: 100%;
}

.form-item label {
  display: block;
  font-size: 16px;
  margin-bottom: 8px;
  color: #333;
}

.form-item input {
  width: 100%;
  height: 50px;
  padding: 0 15px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 16px;
  color: #333;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-item input:focus {
  outline: none;
  border-color: #4DA1FF;
  box-shadow: 0 0 0 1px #4DA1FF;
}

.form-item input:disabled {
  background-color: #f5f7fa;
  cursor: not-allowed;
}

.error-message {
  margin-top: 5px;
  color: #f56c6c;
  font-size: 14px;
}

.password-section-toggle {
  margin: 20px 0;
  display: flex;
  justify-content: flex-start;
}

.toggle-button {
  padding: 8px 16px;
  background-color: transparent;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.toggle-button:hover {
  border-color: #c6e2ff;
  color: #4DA1FF;
  background-color: #ecf5ff;
}

.password-section {
  margin-top: 20px;
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border-left: 3px solid #4DA1FF;
}

.section-title {
  margin-top: 0;
  margin-bottom: 20px;
  font-size: 18px;
  color: #303133;
}

.password-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.password-button {
  padding: 10px 20px;
  background-color: #4DA1FF;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.password-button:hover {
  background-color: #3A90F8;
}

.password-button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.form-actions {
  display: flex;
  justify-content: center;
  gap: 15px;
  margin-top: 30px;
}

.submit-button {
  padding: 12px 24px;
  background-color: #4DA1FF;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s;
  min-width: 120px;
}

.submit-button:hover {
  background-color: #3A90F8;
}

.submit-button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.cancel-button {
  padding: 12px 24px;
  background-color: white;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 120px;
}

.cancel-button:hover {
  border-color: #c6e2ff;
  color: #4DA1FF;
  background-color: #ecf5ff;
}

/* 消息样式 */
:global(.message) {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 20px;
  border-radius: 4px;
  font-size: 14px;
  z-index: 9999;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: fadeIn 0.3s;
}

:global(.message.success) {
  background-color: #f0f9eb;
  border: 1px solid #e1f3d8;
  color: #67c23a;
}

:global(.message.error) {
  background-color: #fef0f0;
  border: 1px solid #fde2e2;
  color: #f56c6c;
}

:global(.message.fade-out) {
  animation: fadeOut 0.3s;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translate(-50%, -20px); }
  to { opacity: 1; transform: translate(-50%, 0); }
}

@keyframes fadeOut {
  from { opacity: 1; transform: translate(-50%, 0); }
  to { opacity: 0; transform: translate(-50%, -20px); }
}

@media (max-width: 768px) {
  .profile-edit-form-container {
    padding: 25px 20px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }

  .back-button {
    align-self: flex-start;
  }

  .form-actions {
    flex-direction: column;
  }

  .submit-button, .cancel-button {
    width: 100%;
  }
}
</style>
