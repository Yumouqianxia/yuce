<template>
  <div class="change-password-container">
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else class="change-password-content">
      <div class="page-header">
        <h2>修改密码</h2>
        <el-button class="blue-btn back-btn" @click="goBack">
          返回个人资料
        </el-button>
      </div>

      <div class="password-card">
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="120px"
          class="password-form"
        >
          <el-form-item label="当前密码" prop="currentPassword">
            <el-input
              v-model="passwordForm.currentPassword"
              type="password"
              show-password
              autocomplete="off"
            />
          </el-form-item>

          <el-form-item label="新密码" prop="newPassword">
            <el-input
              v-model="passwordForm.newPassword"
              type="password"
              show-password
              autocomplete="off"
            />
          </el-form-item>

          <el-form-item label="确认新密码" prop="confirmPassword">
            <el-input
              v-model="passwordForm.confirmPassword"
              type="password"
              show-password
              autocomplete="off"
            />
          </el-form-item>

          <el-form-item class="form-buttons">
            <el-button class="blue-btn" @click="updatePassword" :loading="changingPassword">
              更新密码
            </el-button>
            <el-button @click="goBack">
              取消
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElSkeleton, ElButton, ElForm, ElFormItem, ElInput } from 'element-plus'
import { changePassword } from '@/api/auth'

// 不需要使用userStore
const router = useRouter()
const passwordFormRef = ref<typeof ElForm>()

// 状态
const loading = ref(false)
const changingPassword = ref(false)

// 密码表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码验证规则
const passwordRules = {
  currentPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为6-20个字符', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为6-20个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: (error?: Error) => void) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 更新密码
const updatePassword = async () => {
  try {
    await passwordFormRef.value?.validate()

    changingPassword.value = true
    await changePassword(
      passwordForm.currentPassword,
      passwordForm.newPassword,
      passwordForm.confirmPassword
    )

    // 清空表单
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''

    ElMessage.success('密码更新成功')

    // 更新成功后返回个人资料页
    goBack()
  } catch (error) {
    if (error instanceof Error) {
      ElMessage.error(error.message || '更新密码失败')
    } else {
      ElMessage.error('更新密码失败')
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
onMounted(() => {
  // 这里不需要加载任何数据
  loading.value = false
})
</script>

<style scoped>
.change-password-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 30px 20px;
}

.loading-container {
  padding: 40px 20px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.password-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.password-form {
  max-width: 500px;
  margin: 0 auto;
}

.form-buttons {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}

/* 蓝色按钮样式 */
.blue-btn {
  background-color: #4DA1FF !important;
  border-color: #4DA1FF !important;
  color: white !important;
  border-radius: 4px !important;
  padding: 10px 20px !important;
  font-size: 14px !important;
  transition: all 0.3s ease !important;
  border: none !important;
  font-weight: normal !important;
}

.blue-btn:hover {
  background-color: #3A90F8 !important;
  box-shadow: 0 2px 8px rgba(77, 161, 255, 0.3) !important;
}

.blue-btn:active {
  background-color: #2980F5 !important;
}

@media (max-width: 768px) {
  .change-password-container {
    padding: 20px 15px;
  }

  .password-card {
    padding: 25px 20px;
  }

  .page-header {
    flex-direction: column;
    gap: 15px;
    margin-bottom: 20px;
  }

  .password-form {
    max-width: 100%;
  }
}
</style>
