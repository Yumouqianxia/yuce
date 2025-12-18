<template>
  <div class="admin-section">
    <AdminPermissionCheck permission="user.manage">
      <div class="page-header">
        <h2 class="page-title">用户管理</h2>
        
        <!-- 管理员操作按钮 -->
        <div class="admin-actions">
          <PermissionWrapper permission="admin.manage" :show-fallback="false">
            <el-button type="success" @click="navigateToAdminManagement">
              <el-icon><UserFilled /></el-icon>
              管理员管理
            </el-button>
          </PermissionWrapper>
          
          <PermissionWrapper permission="audit_log.view" :show-fallback="false">
            <el-button type="info" @click="navigateToAuditLogs">
              <el-icon><Document /></el-icon>
              审计日志
            </el-button>
          </PermissionWrapper>
          
          <el-button type="warning" @click="exportUserData">
            <el-icon><Download /></el-icon>
            导出数据
          </el-button>
        </div>
      </div>

      <div class="action-row">
      <div class="search-filters">
        <div class="search-item">
          <label for="id-search">ID</label>
          <input
            id="id-search"
            v-model="idSearch"
            type="text"
            class="input-field"
            @input="handleSearch"
          />
        </div>

        <div class="search-item">
          <label for="name-search">用户名</label>
          <div class="input-with-button">
            <input
              id="name-search"
              v-model="searchQuery"
              type="text"
              class="input-field"
              @input="handleSearch"
            />
            <button class="search-button" @click="handleSearch">
              搜索
            </button>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="roleFilter" class="select-field" @change="handleSearch">
              <option value="">角色</option>
              <option value="user">用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
        </div>

        <button class="add-button" @click="showAddUserDialog = true">
          添加用户
        </button>
      </div>
    </div>

    <AdminUserTable
      :users="currentPageUsers"
      :loading="loading"
      :current-page="currentPage"
      :total-pages="totalPages"
      :page-list="pageList"
      :page-size="pageSize"
      @edit="editUser"
      @delete="deleteUser"
      @reset-password="openPasswordDialog"
      @page-change="handleCurrentChange"
      @page-size-change="handleSizeChange"
    />

    <!-- 编辑用户对话框 -->
    <div v-if="dialogVisible" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ currentUser.id ? '编辑用户' : '添加用户' }}</h3>
          <button class="close-btn" @click="dialogVisible = false">&times;</button>
        </div>
        <form @submit.prevent="saveUser" class="user-form">
          <div class="form-group">
            <label for="username">用户名</label>
            <input
              id="username"
              v-model="currentUser.username"
              type="text"
              class="input-field"
              :disabled="!!currentUser.id"
              required
            />
          </div>

          <div class="form-group">
            <label for="nickname">昵称</label>
            <input
              id="nickname"
              v-model="currentUser.nickname"
              type="text"
              class="input-field"
            />
          </div>

          <div class="form-group">
            <label for="email">邮箱</label>
            <input
              id="email"
              v-model="currentUser.email"
              type="email"
              class="input-field"
              required
            />
          </div>

          <div class="form-group">
            <label for="role">角色</label>
            <select
              id="role"
              v-model="currentUser.role"
              class="select-field"
              :disabled="currentUser.username === 'root'"
            >
              <option value="user">用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>

          <div class="form-group" v-if="!currentUser.id">
            <label for="password">密码</label>
            <input
              id="password"
              v-model="currentUser.password"
              type="password"
              class="input-field"
              required
              minlength="6"
            />
          </div>

          <!-- 添加积分字段 -->
          <div class="form-group" v-if="currentUser.id">
            <label for="points">积分</label>
            <div class="points-edit-container">
              <input
                id="points"
                v-model.number="currentUser.points"
                type="number"
                class="input-field"
                min="0"
              />
              <button type="button" class="btn-history" @click="showPointsHistory(currentUser.id)">查看历史</button>
            </div>
          </div>

          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="dialogVisible = false">取消</button>
            <button type="submit" class="submit-btn">保存</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 添加用户对话框 -->
    <div v-if="showAddUserDialog" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>添加用户</h3>
          <button class="close-btn" @click="showAddUserDialog = false">&times;</button>
        </div>
        <form @submit.prevent="addUser" class="user-form">
          <div class="form-group">
            <label for="new-username">用户名</label>
            <input
              id="new-username"
              v-model="newUser.username"
              type="text"
              class="input-field"
              required
              minlength="3"
              maxlength="20"
            />
          </div>

          <div class="form-group">
            <label for="new-nickname">昵称</label>
            <input
              id="new-nickname"
              v-model="newUser.nickname"
              type="text"
              class="input-field"
            />
          </div>

          <div class="form-group">
            <label for="new-email">邮箱</label>
            <input
              id="new-email"
              v-model="newUser.email"
              type="email"
              class="input-field"
              required
            />
          </div>

          <div class="form-group">
            <label for="new-role">角色</label>
            <select
              id="new-role"
              v-model="newUser.role"
              class="select-field"
            >
              <option value="user">用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>

          <div class="form-group">
            <label for="new-password">密码</label>
            <input
              id="new-password"
              v-model="newUser.password"
              type="password"
              class="input-field"
              required
              minlength="6"
              maxlength="30"
              placeholder="密码长度必须在6-30字符之间"
            />
          </div>

          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="showAddUserDialog = false">取消</button>
            <button type="submit" class="submit-btn">保存</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 重置密码对话框 -->
    <div v-if="passwordDialogVisible" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>重置密码</h3>
          <button class="close-btn" @click="passwordDialogVisible = false">&times;</button>
        </div>
        <form @submit.prevent="submitPasswordChange" class="user-form">
          <div class="form-group">
            <label for="new-password">新密码</label>
            <input
              id="new-password"
              v-model="passwordForm.password"
              type="password"
              class="input-field"
              required
              minlength="6"
              placeholder="至少6位"
            />
          </div>
          <div class="form-group">
            <label for="confirm-password">确认新密码</label>
            <input
              id="confirm-password"
              v-model="passwordForm.confirm"
              type="password"
              class="input-field"
              required
              minlength="6"
              placeholder="再次输入新密码"
            />
          </div>
          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="passwordDialogVisible = false">取消</button>
            <button type="submit" class="submit-btn">确定</button>
          </div>
        </form>
      </div>
    </div>
    </AdminPermissionCheck>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UserFilled, Document, Download } from '@element-plus/icons-vue'
import { User } from '@/types/user'
import AdminPermissionCheck from '@/components/admin/AdminPermissionCheck.vue'
import PermissionWrapper from '@/components/base/PermissionWrapper.vue'
import AdminUserTable from '@/components/admin/AdminUserTable.vue'
import {
  getUsers,
  createUser as createUserApi,
  updateUser as updateUserApi,
  deleteUser as deleteUserApi,
  resetUserPassword as resetUserPasswordApi
} from '@/api/admin'

const router = useRouter()
const userStore = useUserStore()
const users = ref<User[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const showAddUserDialog = ref(false)
const passwordDialogVisible = ref(false)
const idSearch = ref('')
const searchQuery = ref('')
const roleFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalUsers = ref(0)

const currentUser = reactive<{
  id: number | null;
  username: string;
  nickname: string;
  email: string;
  role: string;
  password: string;
  points?: number;
}>({
  id: null,
  username: '',
  nickname: '',
  email: '',
  role: 'user',
  password: '',
  points: 0
})

const newUser = reactive<{
  username: string;
  nickname: string;
  email: string;
  password: string;
  role: string;
}>({
  username: '',
  nickname: '',
  email: '',
  password: '',
  role: 'user'
})

const passwordForm = reactive({
  password: '',
  confirm: ''
})
const passwordUserId = ref<number | null>(null)

// 过滤用户列表
const filteredUsers = computed(() => {
  let result = [...users.value]

  if (idSearch.value) {
    result = result.filter(user =>
      user.id.toString().includes(idSearch.value)
    )
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(user =>
      user.username.toLowerCase().includes(query) ||
      (user.nickname && user.nickname.toLowerCase().includes(query)) ||
      user.email.toLowerCase().includes(query)
    )
  }

  if (roleFilter.value) {
    result = result.filter(user => user.role === roleFilter.value)
  }

  return result
})

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(filteredUsers.value.length / pageSize.value) || 1
})

// 计算要显示的页码列表
const pageList = computed(() => {
  if (totalPages.value <= 7) {
    return Array.from({ length: totalPages.value }, (_, i) => i + 1)
  }

  const pages = []

  if (currentPage.value <= 3) {
    // 当前页靠近开始位置
    for (let i = 1; i <= 5; i++) {
      pages.push(i)
    }
  } else if (currentPage.value >= totalPages.value - 2) {
    // 当前页靠近结束位置
    for (let i = totalPages.value - 4; i <= totalPages.value; i++) {
      pages.push(i)
    }
  } else {
    // 当前页在中间位置
    for (let i = currentPage.value - 2; i <= currentPage.value + 2; i++) {
      pages.push(i)
    }
  }

  return pages
})

// 获取当前页的用户数据
const currentPageUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredUsers.value.slice(start, end)
})

// 获取用户数据
const fetchUsers = async () => {
  loading.value = true
  try {
    const response: any = await getUsers()
    if (response && Array.isArray(response.users)) {
      users.value = response.users
      totalUsers.value = response.total || response.users.length
    } else if (Array.isArray(response)) {
      users.value = response
      totalUsers.value = users.value.length
    } else {
      users.value = []
      totalUsers.value = 0
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 编辑用户
const editUser = (user: User) => {
  Object.assign(currentUser, user)
  currentUser.password = ''
  
  // 获取用户积分
  fetchUserPoints(user.id)
  
  dialogVisible.value = true
}

// 打开重置密码弹窗
const openPasswordDialog = (user: User) => {
  passwordUserId.value = user.id
  passwordForm.password = ''
  passwordForm.confirm = ''
  passwordDialogVisible.value = true
}

// 提交密码重置
const submitPasswordChange = async () => {
  if (!passwordUserId.value) return

  if (!passwordForm.password || passwordForm.password.length < 6) {
    ElMessage({
      message: '密码长度至少6位',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 2000
    })
    return
  }

  if (passwordForm.password !== passwordForm.confirm) {
    ElMessage({
      message: '两次输入的密码不一致',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 2000
    })
    return
  }

  try {
    await resetUserPasswordApi(passwordUserId.value, {
      password: passwordForm.password
    })

    ElMessage({
      message: '密码重置成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 2000
    })

    passwordDialogVisible.value = false
  } catch (error) {
    console.error('重置密码失败:', error)
    ElMessage({
      message: '重置密码失败',
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

// 获取用户积分
const fetchUserPoints = async (userId: number) => {
  // 直接使用用户列表里的 points，避免不存在的积分接口
  const found = users.value.find(u => u.id === userId)
  return found?.points || 0
}

// 添加用户
const addUser = async () => {
  try {
    // 验证表单
    if (typeof validateAddUserForm === 'function' && !validateAddUserForm()) {
      return
    }

    const userData = {
      username: newUser.username,
      password: newUser.password,
      email: newUser.email,
      nickname: newUser.nickname || undefined,
      role: newUser.role
    }

    await createUserApi(userData)

    ElMessage({
      message: '用户添加成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })

    showAddUserDialog.value = false
    fetchUsers()

    // 重置表单
    Object.assign(newUser, {
      username: '',
      nickname: '',
      email: '',
      password: '',
      role: 'user'
    })
  } catch (error) {
    ElMessage.error('添加用户失败，请检查输入信息')
  }
}

// 保存用户
const saveUser = async () => {
  try {
    await updateUserApi(currentUser.id as number, {
      nickname: currentUser.nickname,
      email: currentUser.email,
      role: currentUser.role,
      points: currentUser.points
    })

    ElMessage.success('用户更新成功')
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error('更新用户失败')
  }
}

// 删除用户
const deleteUser = (user: User) => {
  if (user.role === 'admin' || user.username === 'root') {
    ElMessage({
      message: '不能删除管理员账号',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return
  }

  ElMessageBox.confirm(
    `确定要删除用户 "${user.username}" 吗？`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(() => {
      deleteUserConfirmed(user.id)
    })
    .catch(() => {
      ElMessage({
        type: 'info',
        message: '已取消删除',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    })
}

const deleteUserConfirmed = async (userId: number) => {
  try {
    await deleteUserApi(userId)
    ElMessage({
      message: '用户删除成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    fetchUsers()
  } catch (error) {
    console.error('删除用户失败:', error)
    ElMessage({
      message: '删除用户失败',
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

// 跳转到用户积分历史页面
const showPointsHistory = (userId: number | null) => {
  if (userId) {
    // 跳转到正确的用户积分历史页面
    window.open(`/profile/points-history?userId=${userId}`, '_blank')
  }
}

// 搜索与过滤处理
const handleSearch = () => {
  currentPage.value = 1
}

// 分页处理
const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
}

// 验证添加用户表单
const validateAddUserForm = () => {
  // 用户名验证
  if (!newUser.username) {
    ElMessage({
      message: '请输入用户名',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  if (newUser.username.length < 3 || newUser.username.length > 20) {
    ElMessage({
      message: '用户名长度必须在3-20字符之间',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  if (!/^[a-zA-Z0-9]+$/.test(newUser.username)) {
    ElMessage({
      message: '用户名只能包含英文字母和数字',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  // 邮箱验证
  if (!newUser.email) {
    ElMessage({
      message: '请输入邮箱',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  if (!/^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/.test(newUser.email)) {
    ElMessage({
      message: '邮箱格式不正确',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  // 密码验证
  if (!newUser.password) {
    ElMessage({
      message: '请输入密码',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  if (newUser.password.length < 6 || newUser.password.length > 30) {
    ElMessage({
      message: '密码长度必须在6-30字符之间',
      type: 'warning',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    return false
  }

  return true
}

// 导航到管理员管理页面
const navigateToAdminManagement = () => {
  router.push('/admin/admins')
}

// 导航到审计日志页面
const navigateToAuditLogs = () => {
  router.push('/admin/audit-logs')
}

// 导出用户数据
const exportUserData = async () => {
  try {
    ElMessage({
      message: '正在导出用户数据...',
      type: 'info',
      customClass: 'custom-message',
      offset: 80,
      duration: 2000
    })
    
    // 准备导出数据
    const exportData = filteredUsers.value.map(user => ({
      ID: user.id,
      用户名: user.username,
      昵称: user.nickname || '',
      邮箱: user.email,
      角色: user.role === 'admin' ? '管理员' : '用户',
      积分: user.points || 0,
      注册时间: new Date(user.createdAt).toLocaleString()
    }))
    
    // 转换为CSV格式
    const headers = Object.keys(exportData[0] || {})
    const csvContent = [
      headers.join(','),
      ...exportData.map(row => headers.map(header => `"${row[header as keyof typeof row] || ''}"`).join(','))
    ].join('\n')
    
    // 创建下载链接
    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `用户数据_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    ElMessage({
      message: '用户数据导出成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  } catch (error) {
    console.error('导出用户数据失败:', error)
    ElMessage({
      message: '导出用户数据失败',
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

onMounted(() => {
  document.title = '用户管理 | 预测系统'
  fetchUsers()
})
</script>

<style scoped>
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
.admin-section {
  background-color: #fff;
  border-radius: 4px;
  padding: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  margin: 0;
  font-size: 20px;
  color: #303133;
  font-weight: 500;
}

.admin-actions {
  display: flex;
  gap: 12px;
}

.action-row {
  margin-bottom: 20px;
}

.search-filters {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}

.search-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 100px;
}

/* 搜索栏角色下拉框 */
.search-filters .search-item .select-container {
  width: 120px;
}

.search-filters .search-item .select-field {
  width: 120px;
}

.search-item label {
  font-size: 14px;
  color: #606266;
}

.input-field {
  height: 38px;
  padding: 0 15px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  transition: border-color 0.2s;
  outline: none;
  box-sizing: border-box;
}

.input-field:focus {
  border-color: #409eff;
}

.select-container {
  position: relative;
  width: 100%;
  box-sizing: border-box;
}

.select-field {
  height: 38px;
  padding: 0 30px 0 15px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  transition: border-color 0.2s;
  outline: none;
  width: 100%;
  appearance: none;
  box-sizing: border-box;
}

.select-field:focus {
  border-color: #409eff;
}

.select-container::after {
  content: "▼";
  font-size: 12px;
  color: #C0C4CC;
  position: absolute;
  right: 15px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.input-with-button {
  display: flex;
}

.search-button {
  padding: 0 20px;
  height: 38px;
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
  font-size: 14px;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  transition: background-color 0.3s;
}

.search-button:hover {
  background-color: #66b1ff;
}

.add-button {
  padding: 0 20px;
  height: 38px;
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.add-button:hover {
  background-color: #66b1ff;
}

.table-container {
  margin-bottom: 20px;
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #ebeef5;
  text-align: center;
}

.data-table th,
.data-table td {
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
}

.data-table th {
  background-color: #f5f7fa;
  font-weight: 500;
  color: #606266;
}

.data-table tbody tr:nth-child(even) {
  background-color: #fafafa;
}

.data-table tbody tr:hover {
  background-color: #f5f7fa;
}

.loading-cell,
.empty-cell {
  text-align: center;
  padding: 30px;
  color: #909399;
}

.role-tag {
  display: inline-block;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
}

.admin-role {
  background-color: #f56c6c;
  color: #fff;
}

.user-role {
  background-color: #409eff;
  color: #fff;
}

.action-cell {
  white-space: nowrap;
}

.btn-edit,
.btn-delete,
.btn-reset {
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid;
  background-color: transparent;
  font-size: 12px;
  cursor: pointer;
  margin: 0 4px;
}

.btn-reset {
  color: #909399;
  border-color: #dcdfe6;
}

.btn-reset:hover {
  color: #606266;
  border-color: #c0c4cc;
}

.btn-edit {
  color: #409eff;
  border-color: #c6e2ff;
}

.btn-edit:hover {
  color: #fff;
  background-color: #409eff;
  border-color: #409eff;
}

.btn-delete {
  color: #f56c6c;
  border-color: #fbc4c4;
}

.btn-delete:hover {
  color: #fff;
  background-color: #f56c6c;
  border-color: #f56c6c;
}

.btn-delete:disabled {
  color: #c0c4cc;
  border-color: #e4e7ed;
  cursor: not-allowed;
  background-color: #fff;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.pagination-box {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  padding: 8px 10px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.page-numbers {
  display: flex;
  align-items: center;
  margin: 0 10px;
}

.page-number {
  min-width: 32px;
  height: 32px;
  margin: 0 4px;
  padding: 0 4px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  cursor: pointer;
}

.page-number:hover {
  color: #409eff;
}

.page-number.active {
  background-color: #409eff;
  color: #fff;
  border-color: #409eff;
}

.ellipsis {
  display: inline-block;
  width: 24px;
  text-align: center;
  font-weight: bold;
  letter-spacing: 2px;
}

.prev-btn,
.next-btn {
  min-width: 60px;
  height: 32px;
  padding: 0 10px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  cursor: pointer;
}

.prev-btn:hover,
.next-btn:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background-color: #ecf5ff;
}

.prev-btn:disabled,
.next-btn:disabled {
  color: #c0c4cc;
  cursor: not-allowed;
  border-color: #ebeef5;
  background-color: #f4f4f5;
}

.page-size-select {
  display: flex;
  align-items: center;
  margin-left: 15px;
  color: #606266;
  font-size: 14px;
}

.page-size-dropdown {
  margin-left: 5px;
  height: 32px;
  padding: 0 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  outline: none;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 400px;
  max-width: 90%;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #ebeef5;
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.close-btn {
  border: none;
  background: none;
  font-size: 20px;
  color: #909399;
  cursor: pointer;
}

.user-form {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
  box-sizing: border-box;
  width: 100%;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.form-group .input-field,
.form-group .select-field {
  width: 100%;
  box-sizing: border-box;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 25px;
  box-sizing: border-box;
}

.cancel-btn,
.submit-btn {
  padding: 8px 20px;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn {
  background-color: #fff;
  border: 1px solid #dcdfe6;
  color: #606266;
}

.submit-btn {
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
}

.cancel-btn:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background-color: #ecf5ff;
}

.submit-btn:hover {
  background-color: #66b1ff;
  border-color: #66b1ff;
}

/* 积分编辑容器样式 */
.points-edit-container {
  display: flex;
  gap: 10px;
  align-items: center;
}

.btn-history {
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid #c2e7b0;
  background-color: transparent;
  font-size: 12px;
  cursor: pointer;
  color: #67c23a;
}

.btn-history:hover {
  color: #fff;
  background-color: #67c23a;
  border-color: #67c23a;
}

.prediction-options {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}
</style>