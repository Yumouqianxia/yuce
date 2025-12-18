// 管理员UI状态管理
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 管理员页面类型
export type AdminPageType = 
  | 'dashboard'
  | 'admins'
  | 'users'
  | 'sport-types'
  | 'scoring-rules'
  | 'matches'
  | 'predictions'
  | 'audit-logs'
  | 'system'
  | 'cache'

// 弹窗类型
export type DialogType =
  | 'create-admin'
  | 'edit-admin'
  | 'create-sport-type'
  | 'edit-sport-type'
  | 'sport-config'
  | 'create-scoring-rule'
  | 'edit-scoring-rule'
  | 'create-match'
  | 'edit-match'
  | 'match-result'
  | 'user-detail'
  | 'audit-detail'
  | 'permission-manage'
  | 'sport-access-manage'

// 表格列配置
export interface TableColumn {
  key: string
  label: string
  width?: string
  sortable?: boolean
  filterable?: boolean
  visible: boolean
}

export const useAdminUIStore = defineStore('adminUI', () => {
  // ==================== 页面状态 ====================
  
  // 当前活跃页面
  const currentPage = ref<AdminPageType>('dashboard')
  
  // 页面标题
  const pageTitle = ref('管理员控制台')
  
  // 面包屑导航
  const breadcrumbs = ref<Array<{ label: string; path?: string }>>([])
  
  // 侧边栏折叠状态
  const sidebarCollapsed = ref(false)
  
  // 移动端侧边栏显示状态
  const mobileSidebarVisible = ref(false)
  
  // ==================== 弹窗状态 ====================
  
  // 当前打开的弹窗
  const activeDialog = ref<DialogType | null>(null)
  
  // 弹窗数据
  const dialogData = ref<any>(null)
  
  // 弹窗加载状态
  const dialogLoading = ref(false)
  
  // ==================== 表格状态 ====================
  
  // 表格加载状态
  const tableLoading = ref<Record<string, boolean>>({})
  
  // 表格选中项
  const tableSelection = ref<Record<string, any[]>>({})
  
  // 表格排序
  const tableSorting = ref<Record<string, { column: string; order: 'asc' | 'desc' }>>({})
  
  // 表格筛选
  const tableFilters = ref<Record<string, Record<string, any>>>({})
  
  // 表格分页
  const tablePagination = ref<Record<string, { page: number; pageSize: number; total: number }>>({})
  
  // 表格列配置
  const tableColumns = ref<Record<string, TableColumn[]>>({
    admins: [
      { key: 'user_id', label: 'ID', width: '80px', sortable: true, visible: true },
      { key: 'username', label: '用户名', sortable: true, filterable: true, visible: true },
      { key: 'admin_level', label: '管理员级别', sortable: true, filterable: true, visible: true },
      { key: 'is_active', label: '状态', sortable: true, filterable: true, visible: true },
      { key: 'created_at', label: '创建时间', sortable: true, visible: true },
      { key: 'actions', label: '操作', width: '200px', visible: true }
    ],
    sportTypes: [
      { key: 'id', label: 'ID', width: '80px', sortable: true, visible: true },
      { key: 'name', label: '名称', sortable: true, filterable: true, visible: true },
      { key: 'code', label: '代码', sortable: true, filterable: true, visible: true },
      { key: 'category', label: '类别', sortable: true, filterable: true, visible: true },
      { key: 'is_active', label: '状态', sortable: true, filterable: true, visible: true },
      { key: 'sort_order', label: '排序', sortable: true, visible: true },
      { key: 'created_at', label: '创建时间', sortable: true, visible: false },
      { key: 'actions', label: '操作', width: '200px', visible: true }
    ],
    scoringRules: [
      { key: 'id', label: 'ID', width: '80px', sortable: true, visible: true },
      { key: 'name', label: '规则名称', sortable: true, filterable: true, visible: true },
      { key: 'sport_type', label: '运动类型', sortable: true, filterable: true, visible: true },
      { key: 'base_points', label: '基础积分', sortable: true, visible: true },
      { key: 'is_active', label: '状态', sortable: true, filterable: true, visible: true },
      { key: 'created_at', label: '创建时间', sortable: true, visible: false },
      { key: 'actions', label: '操作', width: '200px', visible: true }
    ],
    auditLogs: [
      { key: 'id', label: 'ID', width: '80px', sortable: true, visible: true },
      { key: 'admin_user', label: '操作人', sortable: true, filterable: true, visible: true },
      { key: 'action', label: '操作', sortable: true, filterable: true, visible: true },
      { key: 'resource', label: '资源', sortable: true, filterable: true, visible: true },
      { key: 'status', label: '状态', sortable: true, filterable: true, visible: true },
      { key: 'created_at', label: '操作时间', sortable: true, visible: true },
      { key: 'actions', label: '操作', width: '120px', visible: true }
    ]
  })
  
  // ==================== 搜索状态 ====================
  
  // 搜索关键词
  const searchKeywords = ref<Record<string, string>>({})
  
  // 高级搜索显示状态
  const advancedSearchVisible = ref<Record<string, boolean>>({})
  
  // 高级搜索条件
  const advancedSearchFilters = ref<Record<string, any>>({})
  
  // ==================== 批量操作状态 ====================
  
  // 批量操作模式
  const batchMode = ref<Record<string, boolean>>({})
  
  // 批量操作类型
  const batchAction = ref<Record<string, string>>({})
  
  // ==================== 计算属性 ====================
  
  // 是否有打开的弹窗
  const hasActiveDialog = computed(() => activeDialog.value !== null)
  
  // 当前页面是否有选中项
  const hasSelection = computed(() => {
    const pageKey = currentPage.value
    return tableSelection.value[pageKey]?.length > 0
  })
  
  // 当前页面选中项数量
  const selectionCount = computed(() => {
    const pageKey = currentPage.value
    return tableSelection.value[pageKey]?.length || 0
  })
  
  // 当前页面是否在批量操作模式
  const isInBatchMode = computed(() => {
    const pageKey = currentPage.value
    return batchMode.value[pageKey] || false
  })
  
  // ==================== 页面管理 ====================
  
  // 设置当前页面
  const setCurrentPage = (page: AdminPageType) => {
    currentPage.value = page
    updatePageTitle()
    updateBreadcrumbs()
  }
  
  // 更新页面标题
  const updatePageTitle = () => {
    const titles: Record<AdminPageType, string> = {
      dashboard: '管理员控制台',
      admins: '管理员管理',
      users: '用户管理',
      'sport-types': '运动类型管理',
      'scoring-rules': '积分规则管理',
      matches: '比赛管理',
      predictions: '预测管理',
      'audit-logs': '审计日志',
      system: '系统管理',
      cache: '缓存管理'
    }
    pageTitle.value = titles[currentPage.value] || '管理员控制台'
  }
  
  // 更新面包屑导航
  const updateBreadcrumbs = () => {
    const breadcrumbMap: Record<AdminPageType, Array<{ label: string; path?: string }>> = {
      dashboard: [{ label: '控制台' }],
      admins: [{ label: '控制台', path: '/admin' }, { label: '管理员管理' }],
      users: [{ label: '控制台', path: '/admin' }, { label: '用户管理' }],
      'sport-types': [{ label: '控制台', path: '/admin' }, { label: '运动类型管理' }],
      'scoring-rules': [{ label: '控制台', path: '/admin' }, { label: '积分规则管理' }],
      matches: [{ label: '控制台', path: '/admin' }, { label: '比赛管理' }],
      predictions: [{ label: '控制台', path: '/admin' }, { label: '预测管理' }],
      'audit-logs': [{ label: '控制台', path: '/admin' }, { label: '审计日志' }],
      system: [{ label: '控制台', path: '/admin' }, { label: '系统管理' }],
      cache: [{ label: '控制台', path: '/admin' }, { label: '缓存管理' }]
    }
    breadcrumbs.value = breadcrumbMap[currentPage.value] || [{ label: '控制台' }]
  }
  
  // ==================== 弹窗管理 ====================
  
  // 打开弹窗
  const openDialog = (type: DialogType, data?: any) => {
    activeDialog.value = type
    dialogData.value = data || null
    dialogLoading.value = false
  }
  
  // 关闭弹窗
  const closeDialog = () => {
    activeDialog.value = null
    dialogData.value = null
    dialogLoading.value = false
  }
  
  // 设置弹窗加载状态
  const setDialogLoading = (loading: boolean) => {
    dialogLoading.value = loading
  }
  
  // ==================== 表格管理 ====================
  
  // 设置表格加载状态
  const setTableLoading = (tableKey: string, loading: boolean) => {
    tableLoading.value[tableKey] = loading
  }
  
  // 设置表格选中项
  const setTableSelection = (tableKey: string, selection: any[]) => {
    tableSelection.value[tableKey] = selection
  }
  
  // 清除表格选中项
  const clearTableSelection = (tableKey: string) => {
    tableSelection.value[tableKey] = []
  }
  
  // 设置表格排序
  const setTableSorting = (tableKey: string, column: string, order: 'asc' | 'desc') => {
    tableSorting.value[tableKey] = { column, order }
  }
  
  // 设置表格筛选
  const setTableFilters = (tableKey: string, filters: Record<string, any>) => {
    tableFilters.value[tableKey] = filters
  }
  
  // 设置表格分页
  const setTablePagination = (tableKey: string, pagination: { page: number; pageSize: number; total: number }) => {
    tablePagination.value[tableKey] = pagination
  }
  
  // 获取表格状态
  const getTableState = (tableKey: string) => {
    return {
      loading: tableLoading.value[tableKey] || false,
      selection: tableSelection.value[tableKey] || [],
      sorting: tableSorting.value[tableKey] || null,
      filters: tableFilters.value[tableKey] || {},
      pagination: tablePagination.value[tableKey] || { page: 1, pageSize: 20, total: 0 }
    }
  }
  
  // 重置表格状态
  const resetTableState = (tableKey: string) => {
    delete tableLoading.value[tableKey]
    delete tableSelection.value[tableKey]
    delete tableSorting.value[tableKey]
    delete tableFilters.value[tableKey]
    delete tablePagination.value[tableKey]
  }
  
  // ==================== 搜索管理 ====================
  
  // 设置搜索关键词
  const setSearchKeyword = (pageKey: string, keyword: string) => {
    searchKeywords.value[pageKey] = keyword
  }
  
  // 清除搜索关键词
  const clearSearchKeyword = (pageKey: string) => {
    delete searchKeywords.value[pageKey]
  }
  
  // 切换高级搜索显示状态
  const toggleAdvancedSearch = (pageKey: string) => {
    advancedSearchVisible.value[pageKey] = !advancedSearchVisible.value[pageKey]
  }
  
  // 设置高级搜索条件
  const setAdvancedSearchFilters = (pageKey: string, filters: any) => {
    advancedSearchFilters.value[pageKey] = filters
  }
  
  // 清除高级搜索条件
  const clearAdvancedSearchFilters = (pageKey: string) => {
    delete advancedSearchFilters.value[pageKey]
    advancedSearchVisible.value[pageKey] = false
  }
  
  // ==================== 批量操作管理 ====================
  
  // 进入批量操作模式
  const enterBatchMode = (pageKey: string) => {
    batchMode.value[pageKey] = true
    clearTableSelection(pageKey)
  }
  
  // 退出批量操作模式
  const exitBatchMode = (pageKey: string) => {
    batchMode.value[pageKey] = false
    clearTableSelection(pageKey)
    delete batchAction.value[pageKey]
  }
  
  // 设置批量操作类型
  const setBatchAction = (pageKey: string, action: string) => {
    batchAction.value[pageKey] = action
  }
  
  // ==================== 侧边栏管理 ====================
  
  // 切换侧边栏折叠状态
  const toggleSidebar = () => {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }
  
  // 设置侧边栏折叠状态
  const setSidebarCollapsed = (collapsed: boolean) => {
    sidebarCollapsed.value = collapsed
  }
  
  // 切换移动端侧边栏显示状态
  const toggleMobileSidebar = () => {
    mobileSidebarVisible.value = !mobileSidebarVisible.value
  }
  
  // 设置移动端侧边栏显示状态
  const setMobileSidebarVisible = (visible: boolean) => {
    mobileSidebarVisible.value = visible
  }
  
  // ==================== 表格列管理 ====================
  
  // 设置表格列可见性
  const setColumnVisible = (tableKey: string, columnKey: string, visible: boolean) => {
    const columns = tableColumns.value[tableKey]
    if (columns) {
      const column = columns.find(col => col.key === columnKey)
      if (column) {
        column.visible = visible
      }
    }
  }
  
  // 重置表格列配置
  const resetTableColumns = (tableKey: string) => {
    const columns = tableColumns.value[tableKey]
    if (columns) {
      columns.forEach(column => {
        column.visible = true
      })
    }
  }
  
  // 获取可见的表格列
  const getVisibleColumns = (tableKey: string) => {
    const columns = tableColumns.value[tableKey]
    return columns ? columns.filter(col => col.visible) : []
  }
  
  // ==================== 工具方法 ====================
  
  // 重置所有UI状态
  const resetUIState = () => {
    currentPage.value = 'dashboard'
    pageTitle.value = '管理员控制台'
    breadcrumbs.value = []
    
    activeDialog.value = null
    dialogData.value = null
    dialogLoading.value = false
    
    tableLoading.value = {}
    tableSelection.value = {}
    tableSorting.value = {}
    tableFilters.value = {}
    tablePagination.value = {}
    
    searchKeywords.value = {}
    advancedSearchVisible.value = {}
    advancedSearchFilters.value = {}
    
    batchMode.value = {}
    batchAction.value = {}
  }
  
  return {
    // 页面状态
    currentPage,
    pageTitle,
    breadcrumbs,
    sidebarCollapsed,
    mobileSidebarVisible,
    
    // 弹窗状态
    activeDialog,
    dialogData,
    dialogLoading,
    hasActiveDialog,
    
    // 表格状态
    tableLoading,
    tableSelection,
    tableSorting,
    tableFilters,
    tablePagination,
    tableColumns,
    hasSelection,
    selectionCount,
    
    // 搜索状态
    searchKeywords,
    advancedSearchVisible,
    advancedSearchFilters,
    
    // 批量操作状态
    batchMode,
    batchAction,
    isInBatchMode,
    
    // 页面管理
    setCurrentPage,
    updatePageTitle,
    updateBreadcrumbs,
    
    // 弹窗管理
    openDialog,
    closeDialog,
    setDialogLoading,
    
    // 表格管理
    setTableLoading,
    setTableSelection,
    clearTableSelection,
    setTableSorting,
    setTableFilters,
    setTablePagination,
    getTableState,
    resetTableState,
    
    // 搜索管理
    setSearchKeyword,
    clearSearchKeyword,
    toggleAdvancedSearch,
    setAdvancedSearchFilters,
    clearAdvancedSearchFilters,
    
    // 批量操作管理
    enterBatchMode,
    exitBatchMode,
    setBatchAction,
    
    // 侧边栏管理
    toggleSidebar,
    setSidebarCollapsed,
    toggleMobileSidebar,
    setMobileSidebarVisible,
    
    // 表格列管理
    setColumnVisible,
    resetTableColumns,
    getVisibleColumns,
    
    // 工具方法
    resetUIState
  }
})