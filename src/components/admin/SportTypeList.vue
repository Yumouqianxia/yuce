<template>
  <div class="sport-type-list">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button
          v-if="canManageSportTypes"
          type="primary"
          icon="Plus"
          @click="handleCreate"
        >
          创建运动类型
        </el-button>
        
        <el-button
          v-if="hasSelection"
          type="danger"
          icon="Delete"
          @click="handleBatchDelete"
        >
          批量删除 ({{ selectionCount }})
        </el-button>
        
        <el-button
          icon="Refresh"
          @click="handleRefresh"
        >
          刷新
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索运动类型..."
          style="width: 250px"
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <el-button
          icon="Filter"
          @click="toggleAdvancedSearch"
        >
          高级筛选
        </el-button>
      </div>
    </div>
    
    <!-- 高级搜索 -->
    <el-collapse-transition>
      <div v-show="showAdvancedSearch" class="advanced-search">
        <el-form :model="searchFilters" inline>
          <el-form-item label="运动类别">
            <el-select
              v-model="searchFilters.category"
              placeholder="全部类别"
              clearable
              style="width: 150px"
            >
              <el-option
                v-for="(label, value) in categoryOptions"
                :key="value"
                :label="label"
                :value="value"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="状态">
            <el-select
              v-model="searchFilters.is_active"
              placeholder="全部状态"
              clearable
              style="width: 120px"
            >
              <el-option label="启用" :value="true" />
              <el-option label="禁用" :value="false" />
            </el-select>
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="handleAdvancedSearch">搜索</el-button>
            <el-button @click="handleResetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-collapse-transition>
    
    <!-- 数据表格 -->
    <el-table
      v-loading="loading"
      :data="sportTypes"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column
        v-if="canManageSportTypes"
        type="selection"
        width="55"
      />
      
      <el-table-column
        prop="id"
        label="ID"
        width="80"
        sortable="custom"
      />
      
      <el-table-column
        prop="name"
        label="运动名称"
        min-width="150"
        sortable="custom"
      >
        <template #default="{ row }">
          <div class="sport-name">
            <img
              v-if="row.icon"
              :src="row.icon"
              :alt="row.name"
              class="sport-icon"
            />
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column
        prop="code"
        label="代码"
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.code }}</el-tag>
        </template>
      </el-table-column>
      
      <el-table-column
        prop="category"
        label="类别"
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag
            :type="row.category === 'esports' ? 'primary' : 'success'"
            size="small"
          >
            {{ categoryOptions[row.category as keyof typeof categoryOptions] }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column
        prop="is_active"
        label="状态"
        width="100"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-switch
            v-model="row.is_active"
            :disabled="!canManageSportTypes"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      
      <el-table-column
        prop="sort_order"
        label="排序"
        width="100"
        sortable="custom"
      />
      
      <el-table-column
        label="功能配置"
        width="200"
      >
        <template #default="{ row }">
          <div class="config-tags">
            <el-tag
              v-if="row.configuration?.enable_realtime"
              size="small"
              type="success"
            >
              实时
            </el-tag>
            <el-tag
              v-if="row.configuration?.enable_chat"
              size="small"
              type="primary"
            >
              聊天
            </el-tag>
            <el-tag
              v-if="row.configuration?.enable_voting"
              size="small"
              type="warning"
            >
              投票
            </el-tag>
            <el-tag
              v-if="row.configuration?.enable_prediction"
              size="small"
              type="info"
            >
              预测
            </el-tag>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column
        prop="created_at"
        label="创建时间"
        width="160"
        sortable="custom"
      >
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      
      <el-table-column
        label="操作"
        width="200"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            text
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          
          <el-button
            type="primary"
            size="small"
            text
            @click="handleConfig(row)"
          >
            配置
          </el-button>
          
          <el-button
            v-if="canManageSportTypes"
            type="danger"
            size="small"
            text
            @click="handleDelete(row)"
          >
            删除
          </el-button>
          
          <el-dropdown
            trigger="click"
            @command="(command: string) => handleDropdownCommand(command, row)"
          >
            <el-button type="primary" size="small" text>
              更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="stats">查看统计</el-dropdown-item>
                <el-dropdown-item command="export">导出配置</el-dropdown-item>
                <el-dropdown-item command="copy">复制配置</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
      />
    </div>
    
    <!-- 运动类型管理弹窗 -->
    <SportTypeManager
      v-model="showManager"
      :sport-type="currentSportType"
      @success="handleManagerSuccess"
    />
    
    <!-- 运动配置弹窗 -->
    <SportConfigDialog
      v-model="showConfig"
      :sport-type="currentSportType"
      @success="handleConfigSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Delete, Refresh, Filter, ArrowDown } from '@element-plus/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { usePermissionStore } from '@/stores/permissions'
import { useAdminUIStore } from '@/stores/adminUI'
import { SPORT_CATEGORY_NAMES, SportCategory } from '@/types/admin'
import type { SportType, ListSportTypesRequest } from '@/types/admin'
import SportTypeManager from './SportTypeManager.vue'
import SportConfigDialog from './SportConfigDialog.vue'

const adminStore = useAdminStore()
const permissionStore = usePermissionStore()
const uiStore = useAdminUIStore()

// 状态
const loading = ref(false)
const showManager = ref(false)
const showConfig = ref(false)
const currentSportType = ref<SportType | null>(null)
const searchKeyword = ref('')
const showAdvancedSearch = ref(false)

// 搜索筛选条件
const searchFilters = ref({
  category: '',
  is_active: null as boolean | null
})

// 分页状态
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 计算属性
const sportTypes = computed(() => adminStore.sportTypes)
const canManageSportTypes = computed(() => permissionStore.canManageSportTypes)
const categoryOptions = computed(() => SPORT_CATEGORY_NAMES)

// 表格选择状态
const selection = ref<SportType[]>([])
const hasSelection = computed(() => selection.value.length > 0)
const selectionCount = computed(() => selection.value.length)

// 初始化
onMounted(() => {
  loadSportTypes()
})

// 加载运动类型列表
const loadSportTypes = async () => {
  try {
    loading.value = true
    
    const params: ListSportTypesRequest = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    }
    
    // 添加搜索条件
    if (searchKeyword.value) {
      params.search = searchKeyword.value
    }
    
    if (searchFilters.value.category) {
      params.category = searchFilters.value.category as SportCategory
    }
    
    if (searchFilters.value.is_active !== null) {
      params.is_active = searchFilters.value.is_active
    }
    
    const response = await adminStore.fetchSportTypes(params)
    pagination.value.total = response.total
    
  } catch (error) {
    console.error('加载运动类型列表失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.value.page = 1
  loadSportTypes()
}

// 切换高级搜索
const toggleAdvancedSearch = () => {
  showAdvancedSearch.value = !showAdvancedSearch.value
}

// 高级搜索
const handleAdvancedSearch = () => {
  pagination.value.page = 1
  loadSportTypes()
}

// 重置搜索
const handleResetSearch = () => {
  searchKeyword.value = ''
  searchFilters.value = {
    category: '',
    is_active: null
  }
  pagination.value.page = 1
  loadSportTypes()
}

// 刷新数据
const handleRefresh = () => {
  loadSportTypes()
}

// 创建运动类型
const handleCreate = () => {
  currentSportType.value = null
  showManager.value = true
}

// 编辑运动类型
const handleEdit = (sportType: SportType) => {
  currentSportType.value = sportType
  showManager.value = true
}

// 配置运动类型
const handleConfig = (sportType: SportType) => {
  currentSportType.value = sportType
  showConfig.value = true
}

// 删除运动类型
const handleDelete = async (sportType: SportType) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除运动类型"${sportType.name}"吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await adminStore.deleteSportType(sportType.id)
    ElMessage.success('删除成功')
    loadSportTypes()
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (!hasSelection.value) return
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectionCount.value} 个运动类型吗？此操作不可恢复。`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 批量删除
    const deletePromises = selection.value.map(sportType => 
      adminStore.deleteSportType(sportType.id)
    )
    
    await Promise.all(deletePromises)
    ElMessage.success('批量删除成功')
    loadSportTypes()
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

// 状态切换
const handleStatusChange = async (sportType: SportType) => {
  try {
    await adminStore.updateSportType(sportType.id, {
      is_active: sportType.is_active
    })
    
    ElMessage.success(`${sportType.is_active ? '启用' : '禁用'}成功`)
  } catch (error) {
    // 恢复状态
    sportType.is_active = !sportType.is_active
    console.error('状态切换失败:', error)
    ElMessage.error('状态切换失败')
  }
}

// 表格选择变化
const handleSelectionChange = (newSelection: SportType[]) => {
  selection.value = newSelection
}

// 排序变化
const handleSortChange = ({ column, prop, order }: any) => {
  // 这里可以实现排序逻辑
  console.log('排序变化:', { column, prop, order })
}

// 分页变化
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadSportTypes()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadSportTypes()
}

// 下拉菜单命令
const handleDropdownCommand = (command: string, sportType: SportType) => {
  switch (command) {
    case 'stats':
      // 查看统计
      console.log('查看统计:', sportType)
      break
    case 'export':
      // 导出配置
      console.log('导出配置:', sportType)
      break
    case 'copy':
      // 复制配置
      console.log('复制配置:', sportType)
      break
  }
}

// 管理器成功回调
const handleManagerSuccess = () => {
  loadSportTypes()
}

// 配置成功回调
const handleConfigSuccess = () => {
  loadSportTypes()
}

// 格式化日期时间
const formatDateTime = (dateTime: string) => {
  return new Date(dateTime).toLocaleString('zh-CN')
}
</script>

<style scoped>
.sport-type-list {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.advanced-search {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.sport-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sport-icon {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  object-fit: cover;
}

.config-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

:deep(.el-table) {
  border: 1px solid #EBEEF5;
}

:deep(.el-table th) {
  background-color: #fafafa;
}

:deep(.el-tag) {
  margin: 2px;
}
</style>