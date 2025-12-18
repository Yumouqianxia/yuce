<template>
  <div class="admin-scoring-rules">
    <div class="page-header">
      <h1>积分规则管理</h1>
      <p>管理预测系统的积分计算规则</p>
    </div>

    <div class="content-card">
      <div class="card-header">
        <h2>积分规则配置</h2>
        <el-button type="primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          添加规则
        </el-button>
      </div>

      <div class="table-container">
        <el-table 
          :data="scoringRules" 
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="规则名称" />
          <el-table-column prop="type" label="规则类型" width="120">
            <template #default="scope">
              <el-tag :type="getRuleTypeColor(scope.row.type)">
                {{ getRuleTypeName(scope.row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="points" label="积分值" width="100" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="isActive" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.isActive ? 'success' : 'danger'">
                {{ scope.row.isActive ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editRule(scope.row)">
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
                @click="deleteRule(scope.row)"
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
      :title="editingId ? '编辑积分规则' : '添加积分规则'"
      v-model="showAddDialog"
      width="600px"
    >
      <el-form 
        :model="formData" 
        :rules="formRules" 
        ref="formRef"
        label-width="120px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="完全正确" value="exact" />
            <el-option label="部分正确" value="partial" />
            <el-option label="参与奖励" value="participation" />
            <el-option label="连胜奖励" value="streak" />
            <el-option label="特殊奖励" value="special" />
          </el-select>
        </el-form-item>
        <el-form-item label="积分值" prop="points">
          <el-input-number 
            v-model="formData.points" 
            :min="0" 
            :max="1000"
            placeholder="请输入积分值"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="formData.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入规则描述"
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
          <el-button type="primary" @click="saveRule" :loading="saving">
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

// 积分规则列表
const scoringRules = ref([
  {
    id: 1,
    name: '完全正确预测',
    type: 'exact',
    points: 10,
    description: '预测比赛结果完全正确时获得的积分',
    isActive: true
  },
  {
    id: 2,
    name: '部分正确预测',
    type: 'partial',
    points: 5,
    description: '预测比赛胜负正确但比分不准确时获得的积分',
    isActive: true
  },
  {
    id: 3,
    name: '参与奖励',
    type: 'participation',
    points: 1,
    description: '参与预测即可获得的基础积分',
    isActive: true
  },
  {
    id: 4,
    name: '三连胜奖励',
    type: 'streak',
    points: 15,
    description: '连续三次预测正确的额外奖励积分',
    isActive: true
  },
  {
    id: 5,
    name: '特殊赛事奖励',
    type: 'special',
    points: 20,
    description: '重要赛事（如总决赛）预测正确的额外积分',
    isActive: false
  }
])

// 表单数据
const formData = reactive({
  name: '',
  type: '',
  points: 0,
  description: '',
  isActive: true
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择规则类型', trigger: 'change' }
  ],
  points: [
    { required: true, message: '请输入积分值', trigger: 'blur' },
    { type: 'number', min: 0, message: '积分值不能小于0', trigger: 'blur' }
  ]
}

// 获取规则类型名称
const getRuleTypeName = (type: string) => {
  const typeMap: Record<string, string> = {
    exact: '完全正确',
    partial: '部分正确',
    participation: '参与奖励',
    streak: '连胜奖励',
    special: '特殊奖励'
  }
  return typeMap[type] || type
}

// 获取规则类型颜色
const getRuleTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    exact: 'success',
    partial: 'warning',
    participation: 'info',
    streak: 'primary',
    special: 'danger'
  }
  return colorMap[type] || 'info'
}

// 重置表单
const resetForm = () => {
  formData.name = ''
  formData.type = ''
  formData.points = 0
  formData.description = ''
  formData.isActive = true
  editingId.value = null
}

// 编辑规则
const editRule = (rule: any) => {
  editingId.value = rule.id
  formData.name = rule.name
  formData.type = rule.type
  formData.points = rule.points
  formData.description = rule.description
  formData.isActive = rule.isActive
  showAddDialog.value = true
}

// 保存规则
const saveRule = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    if (editingId.value) {
      // 更新现有规则
      const index = scoringRules.value.findIndex(item => item.id === editingId.value)
      if (index !== -1) {
        scoringRules.value[index] = {
          ...scoringRules.value[index],
          ...formData
        }
      }
      ElMessage.success('积分规则更新成功')
    } else {
      // 添加新规则
      const newRule = {
        id: Date.now(),
        ...formData
      }
      scoringRules.value.unshift(newRule)
      ElMessage.success('积分规则添加成功')
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
const toggleStatus = async (rule: any) => {
  try {
    const action = rule.isActive ? '禁用' : '启用'
    await ElMessageBox.confirm(
      `确定要${action}积分规则"${rule.name}"吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    rule.isActive = !rule.isActive
    ElMessage.success(`${action}成功`)
  } catch (error) {
    // 用户取消操作
  }
}

// 删除规则
const deleteRule = async (rule: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除积分规则"${rule.name}"吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    const index = scoringRules.value.findIndex(item => item.id === rule.id)
    if (index !== -1) {
      scoringRules.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消操作
  }
}

// 加载数据
const loadRules = async () => {
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
  loadRules()
})
</script>

<style scoped>
.admin-scoring-rules {
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