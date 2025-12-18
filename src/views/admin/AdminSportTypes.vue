<template>
  <div class="admin-sport-types">
    <div class="page-header">
      <h1>运动类型管理</h1>
      <p>管理系统中的运动类型和分类</p>
    </div>

    <div class="content-card">
      <div class="card-header">
        <h2>运动类型列表</h2>
        <el-button type="primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          添加运动类型
        </el-button>
      </div>

      <div class="table-container">
        <el-table 
          :data="sportTypes" 
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="运动类型名称" />
          <el-table-column prop="code" label="类型代码" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="isActive" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.isActive ? 'success' : 'danger'">
                {{ scope.row.isActive ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editSportType(scope.row)">
                编辑
              </el-button>
              <el-button 
                size="small" 
                :type="scope.row.isActive ? 'warning' : 'success'"
                @click="toggleStatus(scope.row)"
              >
                {{ scope.row.isActive ? '禁用' : '启用' }}
              </el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteSportType(scope.row)"
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
      :title="editingId ? '编辑运动类型' : '添加运动类型'"
      v-model="showAddDialog"
      width="500px"
    >
      <el-form 
        :model="formData" 
        :rules="formRules" 
        ref="formRef"
        label-width="100px"
      >
        <el-form-item label="类型名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入运动类型名称" />
        </el-form-item>
        <el-form-item label="类型代码" prop="code">
          <el-input v-model="formData.code" placeholder="请输入类型代码（如：football）" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="formData.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入运动类型描述"
          />
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
          <el-button type="primary" @click="saveSportType" :loading="saving">
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

// 运动类型列表
const sportTypes = ref([
  {
    id: 1,
    name: '足球',
    code: 'football',
    description: '世界第一运动，包括各种足球比赛',
    isActive: true,
    createdAt: '2024-01-01T00:00:00Z'
  },
  {
    id: 2,
    name: '篮球',
    code: 'basketball',
    description: 'NBA、CBA等篮球比赛',
    isActive: true,
    createdAt: '2024-01-01T00:00:00Z'
  },
  {
    id: 3,
    name: '网球',
    code: 'tennis',
    description: '温网、法网等网球赛事',
    isActive: false,
    createdAt: '2024-01-01T00:00:00Z'
  }
])

// 表单数据
const formData = reactive({
  name: '',
  code: '',
  description: '',
  isActive: true
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入运动类型名称', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入类型代码', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '代码只能包含小写字母和下划线', trigger: 'blur' }
  ]
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 重置表单
const resetForm = () => {
  formData.name = ''
  formData.code = ''
  formData.description = ''
  formData.isActive = true
  editingId.value = null
}

// 编辑运动类型
const editSportType = (sportType: any) => {
  editingId.value = sportType.id
  formData.name = sportType.name
  formData.code = sportType.code
  formData.description = sportType.description
  formData.isActive = sportType.isActive
  showAddDialog.value = true
}

// 保存运动类型
const saveSportType = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    if (editingId.value) {
      // 更新现有运动类型
      const index = sportTypes.value.findIndex(item => item.id === editingId.value)
      if (index !== -1) {
        sportTypes.value[index] = {
          ...sportTypes.value[index],
          ...formData
        }
      }
      ElMessage.success('运动类型更新成功')
    } else {
      // 添加新运动类型
      const newSportType = {
        id: Date.now(),
        ...formData,
        createdAt: new Date().toISOString()
      }
      sportTypes.value.unshift(newSportType)
      ElMessage.success('运动类型添加成功')
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
const toggleStatus = async (sportType: any) => {
  try {
    const action = sportType.isActive ? '禁用' : '启用'
    await ElMessageBox.confirm(
      `确定要${action}运动类型"${sportType.name}"吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    sportType.isActive = !sportType.isActive
    ElMessage.success(`${action}成功`)
  } catch (error) {
    // 用户取消操作
  }
}

// 删除运动类型
const deleteSportType = async (sportType: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除运动类型"${sportType.name}"吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    const index = sportTypes.value.findIndex(item => item.id === sportType.id)
    if (index !== -1) {
      sportTypes.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消操作
  }
}

// 加载数据
const loadSportTypes = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    // 数据已经在上面定义了，这里只是模拟加载过程
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadSportTypes()
})

// 监听对话框关闭
const handleDialogClose = () => {
  resetForm()
  formRef.value?.resetFields()
}
</script>

<style scoped>
.admin-sport-types {
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
</style>