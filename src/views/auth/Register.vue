<template>
  <div class="register-container">
    <div class="register-form-container">
      <h2 class="title">注册账号</h2>

      <form @submit.prevent="handleRegister" class="form">
        <div class="form-item">
          <label>用户名</label>
          <input
            v-model="registerData.username"
            type="text"
            placeholder="请输入用户名（仅限英文和数字）"
          />
        </div>

        <div class="form-item">
          <label>昵称</label>
          <input
            v-model="registerData.nickname"
            type="text"
            placeholder="请输入昵称（可选）"
          />
        </div>

        <div class="form-item">
          <label>邮箱</label>
          <input
            v-model="registerData.email"
            type="email"
            placeholder="请输入邮箱"
          />
        </div>

        <div class="form-item">
          <label>密码</label>
          <input
            v-model="registerData.password"
            type="password"
            placeholder="请输入密码（6-30位）"
          />
        </div>

        <div class="form-item">
          <label>确认密码</label>
          <input
            v-model="registerData.password_confirm"
            type="password"
            placeholder="请再次输入密码"
          />
        </div>

        <button
          type="submit"
          class="submit-button"
          :disabled="loading"
        >
          {{ loading ? '注册中...' : '注册' }}
        </button>

        <div class="login-link">
          已有账号？<router-link to="/login">立即登录</router-link>
        </div>

        <div class="home-link">
          <router-link to="/">返回首页</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { RegisterData } from '@/types/user'
import { checkUsernameAvailability } from '@/api/auth'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

// 辅助函数：显示消息
const showMessage = (message: string, type: 'success' | 'warning' | 'error') => {
  return ElMessage({
    message,
    type,
    customClass: 'custom-message',
    offset: 80,
    duration: 3000
  })
}

// 注册表单数据
const registerData = reactive<RegisterData>({
  username: '',
  nickname: '',
  email: '',
  password: '',
  password_confirm: ''
})

// 验证表单
const validateForm = async () => {
  // 用户名验证
  if (!registerData.username) {
    showMessage('请输入用户名', 'warning')
    return false
  }

  if (registerData.username.length < 3 || registerData.username.length > 20) {
    showMessage('用户名长度应在3到20个字符之间', 'warning')
    return false
  }

  if (!/^[a-zA-Z0-9]+$/.test(registerData.username)) {
    showMessage('用户名只能包含英文字母和数字', 'warning')
    return false
  }

  try {
    const response = await checkUsernameAvailability(registerData.username)
    if (!response.available) {
      showMessage('该用户名已被占用', 'warning')
      return false
    }
  } catch (error) {
    console.error('检查用户名可用性失败', error)
  }

  // 邮箱验证
  if (!registerData.email) {
    showMessage('请输入邮箱', 'warning')
    return false
  }

  if (!/^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/.test(registerData.email)) {
    showMessage('邮箱格式不正确', 'warning')
    return false
  }

  // 密码验证
  if (!registerData.password) {
    showMessage('请输入密码', 'warning')
    return false
  }

  if (registerData.password.length < 6 || registerData.password.length > 30) {
    showMessage('密码长度必须在6-30字符之间', 'warning')
    return false
  }

  // 确认密码
  if (!registerData.password_confirm) {
    showMessage('请再次输入密码', 'warning')
    return false
  }

  if (registerData.password_confirm !== registerData.password) {
    showMessage('两次输入密码不一致', 'warning')
    return false
  }

  return true
}

// 处理注册
const handleRegister = async () => {
  const isValid = await validateForm()
  if (!isValid) return

  loading.value = true

  try {
    await userStore.register(registerData)
    showMessage('注册成功，请登录', 'success')
    router.push('/login')
  } catch (error) {
    if (error instanceof Error) {
      showMessage(error.message || '注册失败', 'error')
    } else {
      showMessage('注册失败，请检查注册信息', 'error')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 20px;
}

.register-form-container {
  width: 100%;
  max-width: 450px;
  padding: 40px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.title {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
  font-size: 24px;
  font-weight: 500;
}

.form {
  width: 100%;
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
  border-color: #5ecbc9;
  box-shadow: 0 0 0 1px #5ecbc9;
}

.error-message {
  display: none;
}

:global(.custom-message) {
  min-width: 240px;
  padding: 15px 20px;
  display: flex;
  align-items: center;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
}

.submit-button {
  width: 100%;
  height: 50px;
  margin-top: 10px;
  background-color: #5ecbc9;
  border: 1px solid #5ecbc9;
  border-radius: 4px;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.submit-button:hover {
  background-color: #4ebfbd;
}

.submit-button:disabled {
  background-color: #a0cfce;
  cursor: not-allowed;
}

.login-link {
  margin-top: 20px;
  text-align: center;
  font-size: 14px;
  color: #606266;
}

.login-link a {
  color: #5ecbc9;
  text-decoration: none;
}

.home-link {
  margin-top: 12px;
  text-align: center;
  font-size: 14px;
}

.home-link a {
  color: #909399;
  text-decoration: none;
}
</style>