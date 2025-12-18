<template>
  <div class="admin-permission-check">
    <!-- 权限检查通过时显示内容 -->
    <div v-if="hasPermission" class="permission-granted">
      <slot />
    </div>
    
    <!-- 权限检查失败时显示的内容 -->
    <div v-else class="permission-denied">
      <slot name="denied">
        <el-alert
          :title="deniedTitle"
          :description="deniedMessage"
          type="warning"
          :show-icon="true"
          :closable="false"
          class="permission-alert"
        />
      </slot>
    </div>
    
    <!-- 加载状态 -->
    <div v-if="loading" class="permission-loading">
      <el-skeleton :rows="3" animated />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { usePermissionStore } from '@/stores/permissions'
import { useAdminStore } from '@/stores/admin'
import { ADMIN_PERMISSIONS } from '@/types/admin'

interface Props {
  // 需要检查的权限代码
  permission?: string
  // 多个权限（需要全部拥有）
  permissions?: string[]
  // 多个权限（拥有任一即可）
  anyPermissions?: string[]
  // 是否需要管理员身份
  requireAdmin?: boolean
  // 是否需要超级管理员身份
  requireSuperAdmin?: boolean
  // 运动类型ID（用于运动类型访问权限检查）
  sportTypeId?: number
  // 自定义拒绝访问标题
  deniedTitle?: string
  // 自定义拒绝访问消息
  deniedMessage?: string
  // 是否自动初始化权限数据
  autoInit?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  requireAdmin: true,
  requireSuperAdmin: false,
  deniedTitle: '权限不足',
  deniedMessage: '您没有权限访问此功能，请联系管理员',
  autoInit: true
})

const permissionStore = usePermissionStore()
const adminStore = useAdminStore()

// 计算权限检查结果
const hasPermission = computed(() => {
  // 如果正在加载，暂时返回false
  if (loading.value) {
    return false
  }
  
  // 超级管理员检查
  if (props.requireSuperAdmin) {
    return permissionStore.isSuperAdmin
  }
  
  // 管理员身份检查
  if (props.requireAdmin) {
    if (!permissionStore.isAdmin) {
      return false
    }
  }
  
  // 运动类型访问权限检查
  if (props.sportTypeId !== undefined) {
    if (!permissionStore.canAccessSportType(props.sportTypeId)) {
      return false
    }
  }
  
  // 单个权限检查
  if (props.permission) {
    return permissionStore.hasPermission(props.permission)
  }
  
  // 多个权限检查（需要全部拥有）
  if (props.permissions && props.permissions.length > 0) {
    return permissionStore.hasAllPermissions(props.permissions)
  }
  
  // 多个权限检查（拥有任一即可）
  if (props.anyPermissions && props.anyPermissions.length > 0) {
    return permissionStore.hasAnyPermission(props.anyPermissions)
  }
  
  // 如果只要求管理员身份且已通过检查，则允许访问
  if (props.requireAdmin && permissionStore.isAdmin) {
    return true
  }
  
  // 默认拒绝访问
  return false
})

// 加载状态
const loading = computed(() => {
  return permissionStore.permissionsLoading || adminStore.adminLoading
})

// 动态生成拒绝访问消息
const deniedMessage = computed(() => {
  if (props.deniedMessage !== '您没有权限访问此功能，请联系管理员') {
    return props.deniedMessage
  }
  
  // 根据权限类型生成具体消息
  if (props.requireSuperAdmin) {
    return '此功能需要超级管理员权限'
  }
  
  if (props.requireAdmin && !permissionStore.isAdmin) {
    return '此功能需要管理员权限'
  }
  
  if (props.permission) {
    return permissionStore.getPermissionDeniedMessage(props.permission)
  }
  
  if (props.sportTypeId !== undefined) {
    return '您没有访问此运动类型的权限'
  }
  
  return props.deniedMessage
})

// 组件挂载时初始化权限数据
onMounted(async () => {
  if (props.autoInit && permissionStore.isAdmin) {
    try {
      await permissionStore.initializePermissions()
    } catch (error) {
      console.error('初始化权限数据失败:', error)
    }
  }
})

// 暴露权限检查方法供父组件使用
defineExpose({
  hasPermission,
  loading,
  checkPermission: (permission: string) => permissionStore.hasPermission(permission),
  checkPermissions: (permissions: string[]) => permissionStore.hasAllPermissions(permissions),
  checkAnyPermissions: (permissions: string[]) => permissionStore.hasAnyPermission(permissions)
})
</script>

<style scoped>
.admin-permission-check {
  width: 100%;
}

.permission-granted {
  width: 100%;
}

.permission-denied {
  width: 100%;
}

.permission-loading {
  width: 100%;
  padding: 20px;
}

.permission-alert {
  margin: 16px 0;
}

/* 紧凑模式 */
.admin-permission-check.compact .permission-alert {
  margin: 8px 0;
}

.admin-permission-check.compact .permission-loading {
  padding: 10px;
}

/* 内联模式 */
.admin-permission-check.inline {
  display: inline-block;
  width: auto;
}

.admin-permission-check.inline .permission-granted {
  display: inline-block;
  width: auto;
}

.admin-permission-check.inline .permission-denied {
  display: inline-block;
  width: auto;
}

.admin-permission-check.inline .permission-alert {
  margin: 0;
  display: inline-block;
}
</style>