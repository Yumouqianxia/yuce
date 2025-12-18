<template>
  <div class="admin-management">
    <div class="page-header">
      <h1>管理员管理</h1>
      <p>管理系统管理员账户和权限</p>
    </div>

    <div class="content-card">
      <div class="card-header">
        <h2>管理员列表</h2>
        <el-button type="primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          添加管理员
        </el-button>
      </div>

      <div class="table-container">
        <el-table 
          :data="admins" 
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="role" label="角色" width="120">
            <template #default="scope">
              <el-tag :type="getRoleColor(scope.row.role)">
                {{ getRoleName(scope.row.role) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="permissions" label="权限" width="200">
            <template #default="scope">
              <div class="permissions-tags">
                <el-tag 
                  v-for="permission in scope.row.permissions.slice(0, 2)" 
                  :key="permission"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px;"
                >
                  {{ getPermissionName(permission) }}
                </el-tag>
                <el-tag 
                  v-if="scope.row.permissions.length > 2"
                  size="small"
                  type="info"
                >
                  +{{ scope.row.permissions.length - 2 }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="isActive" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.isActive ? 'success' : 'danger'">
                {{ scope.row.isActive ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastLogin" label="最后登录" width="180">
            <template #default="scope">
              {{ scope.row.lastLogin ? formatDate(scope.row.lastLogin) : '从未登录' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editAdmin(scope.row)">
                编辑
              </el-button>
              <el-button 
                size="small" 
                :type="scope.row.isActive ? 'warning' : 'success'"
                @click="toggleStatus(scope.row)"
                :disabled="scope.row.role === 'super_admin'"
              >
                {{ scope.row.isActive ? '禁用' : '启用' }}
              </el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteAdmin(scope.row)"
                :disabled="scope.row.role === 'super_admin'"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog 
      :title="editingId ? '编辑管理员' : '添加管理员'"
      v-model="showAddDialog"
      width="600px"
    >
      <el-form 
        :model="formData" 
        :rules="formRules" 
        ref="formRef"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!editingId">
          <el-input 
            v-model="formData.password" 
            type="password" 
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="运营人员" value="operator" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="formData.permissions">
            <el-checkbox label="user_manage">用户管理</el-checkbox>
            <el-checkbox label="match_manage">比赛管理</el-checkbox>
            <el-checkbox label="site_manage">公告管理</el-checkbox>
            <el-checkbox label="scoring_manage">积分管理</el-checkbox>
            <el-checkbox label="admin_manage">管理员管理</el-checkbox>
            <el-checkbox label="system_settings">系统设置</el-checkbox>
            <el-checkbox label="audit_logs">审计日志</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="状态" prop="isActive">
          <el-switch 
            v-model="formData.isActive"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="saveAdmin" :loading="saving">
            {{ editingId ? '更新' : '添加' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 响应式数据
const loading = ref(false)
const saving = ref(false)
const showAddDialog = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref()

// 管理员列表
const admins = ref([
  {
    id: 1,
    username: 'superadmin',
    email: 'admin@example.com',
    role: 'super_admin',
    permissions: ['user_manage', 'match_manage', 'site_manage', 'scoring_manage', 'admin_manage', 'system_settings', 'audit_logs'],
    isActive: true,
    lastLogin: '2024-01-15T10:30:00Z'
  },
  {
    id: 2,
    username: 'admin1',
    email: 'admin1@example.com',
    role: 'admin',
    permissions: ['user_manage', 'match_manage', 'site_manage'],
    isActive: true,
    lastLogin: '2024-01-14T15:20:00Z'
  },
  {
    id: 3,
    username: 'operator1',
    email: 'operator1@example.com',
    role: 'operator',
    permissions: ['match_manage', 'site_manage'],
    isActive: true,
    lastLogin: null
  }
])

// 表单数据
const formData = reactive({
  username: '',
  email: '',
  password: '',
  role: '',
  permissions: [] as string[],
  isActive: true
})

// 表单验证规则
const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 获取角色名称
const getRoleName = (role: string) => {
  const roleMap: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    operator: '运营人员'
  }
  return roleMap[role] || role
}

// 获取角色颜色
const getRoleColor = (role: string) => {
  const colorMap: Record<string, string> = {
    super_admin: 'danger',
    admin: 'primary',
    operator: 'success'
  }
  return colorMap[role] || 'info'
}

// 获取权限名称
const getPermissionName = (permission: string) => {
  const permissionMap: Record<string, string> = {
    user_manage: '用户管理',
    match_manage: '比赛管理',
    site_manage: '公告管理',
    scoring_manage: '积分管理',
    admin_manage: '管理员管理',
    system_settings: '系统设置',
    audit_logs: '审计日志'
  }
  return permissionMap[permission] || permission
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 重置表单
const resetForm = () => {
  formData.username = ''
  formData.email = ''
  formData.password = ''
  formData.role = ''
  formData.permissions = []
  formData.isActive = true
  editingId.value = null
}

// 编辑管理员
const editAdmin = (admin: any) => {
  editingId.value = admin.id
  formData.username = admin.username
  formData.email = admin.email
  formData.role = admin.role
  formData.permissions = [...admin.permissions]
  formData.isActive = admin.isActive
  showAddDialog.value = true
}

// 保存管理员
const saveAdmin = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    if (editingId.value) {
      // 更新现有管理员
      const index = admins.value.findIndex(item => item.id === editingId.value)
      if (index !== -1) {
        admins.value[index] = {
          ...admins.value[index],
          username: formData.username,
          email: formData.email,
          role: formData.role,
          permissions: [...formData.permissions],
          isActive: formData.isActive
        }
      }
      ElMessage.success('管理员信息更新成功')
    } else {
      // 添加新管理员
      const newAdmin = {
        id: Date.now(),
        username: formData.username,
        email: formData.email,
        role: formData.role,
        permissions: [...formData.permissions],
        isActive: formData.isActive,
        lastLogin: null
      }
      admins.value.unshift(newAdmin)
      ElMessage.success('管理员添加成功')
    }
    
    showAddDialog.value = false
    resetForm()
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

// 切换状态
const toggleStatus = async (admin: any) => {
  try {
    const action = admin.isActive ? '禁用' : '启用'
    await ElMessageBox.confirm(
      `确定要${action}管理员"${admin.username}"吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    admin.isActive = !admin.isActive
    ElMessage.success(`${action}成功`)
  } catch (error) {
    // 用户取消操作
  }
}

// 删除管理员
const deleteAdmin = async (admin: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除管理员"${admin.username}"吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    const index = admins.value.findIndex(item => item.id === admin.id)
    if (index !== -1) {
      admins.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消操作
  }
}

// 加载数据
const loadAdmins = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadAdmins()
})
</script>

<style scoped>
.admin-management {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

.content-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #ebeef5;
}

.card-header h2 {
  margin: 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.table-container {
  padding: 0;
}

.permissions-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-table) {
  border: none;
}

:deep(.el-table th) {
  background-color: #fafafa;
  color: #606266;
  font-weight: 600;
}

:deep(.el-table td) {
  border-bottom: 1px solid #ebeef5;
}

:deep(.el-table tr:hover > td) {
  background-color: #f5f7fa;
}

:deep(.el-checkbox-group) {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>