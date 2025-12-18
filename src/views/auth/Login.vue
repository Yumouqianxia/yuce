<template>
  <div class="login-container">
    <div class="login-form-container">
      <h2 class="title">登录账号</h2>
      
      <form @submit.prevent="handleLogin" class="form">
        <div class="form-item">
          <label>用户名</label>
          <input 
            v-model="loginData.username" 
            type="text"
            placeholder="请输入用户名"
          />
        </div>
        
        <div class="form-item">
          <label>密码</label>
          <input 
            v-model="loginData.password" 
            type="password"
            placeholder="请输入密码" 
          />
        </div>
        
        <button 
          type="submit" 
          class="submit-button"
          :disabled="loading"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
        
        <div class="register-link">
          还没有账号？ <router-link to="/register">立即注册</router-link>
        </div>
        
        <div class="home-link">
          <router-link to="/">返回首页</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { LoginData } from '@/types/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)

// 检查是否已经登录
onMounted(() => {
  if (userStore.isAuthenticated) {
    const redirect = route.query.redirect as string || '/'
    router.replace(redirect)
  }
})

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

// 登录表单数据
const loginData = reactive<LoginData>({
  username: '',
  password: ''
})

// 验证表单
const validateForm = () => {
  // 用户名验证
  if (!loginData.username) {
    showMessage('请输入用户名', 'warning')
    return false
  } 
  
  if (loginData.username.length < 3 || loginData.username.length > 20) {
    showMessage('用户名长度应在3到20个字符之间', 'warning')
    return false
  }
  
  // 密码验证
  if (!loginData.password) {
    showMessage('请输入密码', 'warning')
    return false
  } 
  
  if (loginData.password.length < 8) {
    showMessage('密码长度不得小于8位!', 'warning')
    return false
  }
  
  return true
}

// 处理登录
const handleLogin = async () => {
  if (!validateForm()) return
  
  loading.value = true
  
  try {
    console.log('开始登录流程...')
    await userStore.login(loginData)
    
    // 登录成功后检查认证状态
    if (userStore.isAuthenticated) {
      console.log('登录成功，认证状态：', userStore.isAuthenticated)
      showMessage('登录成功', 'success')
      
      // 延迟一点跳转以便消息显示
      setTimeout(() => {
        // 获取重定向URL，如果存在则跳转到该URL，否则跳转到首页
        const redirect = route.query.redirect as string || '/'
        console.log('跳转到：', redirect)
        router.replace(redirect)
      }, 500)
    } else {
      console.error('登录成功但认证状态为false')
      showMessage('登录状态异常，请刷新页面后重试', 'error')
    }
  } catch (error) {
    console.error('登录错误：', error)
    if (error instanceof Error) {
      showMessage(error.message || '登录失败，请检查用户名和密码', 'error')
    } else {
      showMessage('登录失败，请检查用户名和密码', 'error')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 20px;
}

.login-form-container {
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

.register-link {
  margin-top: 20px;
  text-align: center;
  font-size: 14px;
  color: #606266;
}

.register-link a {
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