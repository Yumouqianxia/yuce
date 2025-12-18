<template>
  <div class="admin-audit-logs">
    <div class="page-header">
      <h1>审计日志</h1>
      <p>查看系统操作记录和安全审计信息</p>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-card">
      <el-form :model="filterForm" inline>
        <el-form-item label="操作类型">
          <el-select v-model="filterForm.action" placeholder="全部" clearable style="width: 150px">
            <el-option label="登录" value="login" />
            <el-option label="登出" value="logout" />
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="查看" value="view" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作模块">
          <el-select v-model="filterForm.module" placeholder="全部" clearable style="width: 150px">
            <el-option label="用户管理" value="user" />
            <el-option label="比赛管理" value="match" />
            <el-option label="公告管理" value="site" />
            <el-option label="积分管理" value="scoring" />
            <el-option label="管理员" value="admin" />
            <el-option label="系统设置" value="system" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="filterForm.username" placeholder="请输入用户名" style="width: 150px" />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchLogs">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetFilter">重置</el-button>
          <el-button type="success" @click="exportLogs">
            <el-icon><Download /></el-icon>
            导出
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="content-card">
      <div class="card-header">
        <h2>操作日志</h2>
        <div class="header-actions">
          <el-button size="small" @click="refreshLogs">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <div class="table-container">
        <el-table 
          :data="auditLogs" 
          v-loading="loading"
          style="width: 100%"
          :default-sort="{ prop: 'createdAt', order: 'descending' }"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="操作人" width="120" />
          <el-table-column prop="action" label="操作类型" width="100">
            <template #default="scope">
              <el-tag :type="getActionColor(scope.row.action)">
                {{ getActionName(scope.row.action) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="module" label="操作模块" width="120">
            <template #default="scope">
              <el-tag type="info">
                {{ getModuleName(scope.row.module) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="操作描述" min-width="200" />
          <el-table-column prop="ip" label="IP地址" width="140" />
          <el-table-column prop="userAgent" label="用户代理" width="200" show-overflow-tooltip />
          <el-table-column prop="createdAt" label="操作时间" width="180" sortable>
            <template #default="scope">
              {{ formatDate(scope.row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button size="small" @click="viewDetails(scope.row)">
                详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.size"
            :page-sizes="[10, 20, 50, 100]"
            :total="pagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>
    </div>

    <!-- 详情对话框 -->
    <el-dialog 
      title="操作详情"
      v-model="showDetailDialog"
      width="600px"
    >
      <div v-if="selectedLog" class="log-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="操作ID">{{ selectedLog.id }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ selectedLog.username }}</el-descriptions-item>
          <el-descriptions-item label="操作类型">
            <el-tag :type="getActionColor(selectedLog.action)">
              {{ getActionName(selectedLog.action) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="操作模块">
            <el-tag type="info">
              {{ getModuleName(selectedLog.module) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="IP地址">{{ selectedLog.ip }}</el-descriptions-item>
          <el-descriptions-item label="操作时间">{{ formatDate(selectedLog.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="操作描述" :span="2">{{ selectedLog.description }}</el-descriptions-item>
          <el-descriptions-item label="用户代理" :span="2">{{ selectedLog.userAgent }}</el-descriptions-item>
        </el-descriptions>
        
        <div v-if="selectedLog.details" class="log-details-section">
          <h4>详细信息</h4>
          <el-input
            v-model="selectedLog.details"
            type="textarea"
            :rows="6"
            readonly
          />
        </div>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showDetailDialog = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download, Refresh } from '@element-plus/icons-vue'

// 响应式数据
const loading = ref(false)
const showDetailDialog = ref(false)
const selectedLog = ref<any>(null)

// 筛选表单
const filterForm = reactive({
  action: '',
  module: '',
  username: '',
  dateRange: null as any
})

// 分页数据
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 审计日志列表
const auditLogs = ref([
  {
    id: 1,
    username: 'admin',
    action: 'login',
    module: 'system',
    description: '管理员登录系统',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15T10:30:00Z',
    details: '{"loginMethod": "password", "success": true, "sessionId": "sess_123456"}'
  },
  {
    id: 2,
    username: 'admin',
    action: 'create',
    module: 'user',
    description: '创建新用户: testuser',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15T10:35:00Z',
    details: '{"userId": 123, "username": "testuser", "email": "test@example.com"}'
  },
  {
    id: 3,
    username: 'admin',
    action: 'update',
    module: 'match',
    description: '更新比赛信息: KPL春季赛第1轮',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15T11:00:00Z',
    details: '{"matchId": 456, "changes": {"status": "completed", "score": "2-1"}}'
  },
  {
    id: 4,
    username: 'operator1',
    action: 'create',
    module: 'site',
    description: '发布新公告: 系统维护通知',
    ip: '192.168.1.101',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    createdAt: '2024-01-15T14:20:00Z',
    details: '{"announcementId": 789, "title": "系统维护通知", "type": "maintenance"}'
  },
  {
    id: 5,
    username: 'admin',
    action: 'delete',
    module: 'user',
    description: '删除用户: spamuser',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15T15:45:00Z',
    details: '{"userId": 999, "username": "spamuser", "reason": "违规行为"}'
  }
])

// 获取操作类型名称
const getActionName = (action: string) => {
  const actionMap: Record<string, string> = {
    login: '登录',
    logout: '登出',
    create: '创建',
    update: '更新',
    delete: '删除',
    view: '查看'
  }
  return actionMap[action] || action
}

// 获取操作类型颜色
const getActionColor = (action: string) => {
  const colorMap: Record<string, string> = {
    login: 'success',
    logout: 'info',
    create: 'primary',
    update: 'warning',
    delete: 'danger',
    view: 'info'
  }
  return colorMap[action] || 'info'
}

// 获取模块名称
const getModuleName = (module: string) => {
  const moduleMap: Record<string, string> = {
    user: '用户管理',
    match: '比赛管理',
    site: '公告管理',
    scoring: '积分管理',
    admin: '管理员',
    system: '系统设置'
  }
  return moduleMap[module] || module
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 搜索日志
const searchLogs = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('搜索完成')
  } catch (error) {
    ElMessage.error('搜索失败')
  } finally {
    loading.value = false
  }
}

// 重置筛选
const resetFilter = () => {
  filterForm.action = ''
  filterForm.module = ''
  filterForm.username = ''
  filterForm.dateRange = null
  searchLogs()
}

// 导出日志
const exportLogs = async () => {
  try {
    ElMessage.success('导出功能开发中...')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

// 刷新日志
const refreshLogs = () => {
  loadLogs()
}

// 查看详情
const viewDetails = (log: any) => {
  selectedLog.value = log
  showDetailDialog.value = true
}

// 处理页面大小变化
const handleSizeChange = (size: number) => {
  pagination.size = size
  loadLogs()
}

// 处理当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  loadLogs()
}

// 加载日志数据
const loadLogs = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    pagination.total = 100 // 模拟总数
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.admin-audit-logs {
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

.filter-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
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

.header-actions {
  display: flex;
  gap: 10px;
}

.table-container {
  padding: 0;
}

.pagination-container {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.log-detail {
  padding: 10px 0;
}

.log-details-section {
  margin-top: 20px;
}

.log-details-section h4 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 14px;
  font-weight: 600;
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