<template>
  <div v-if="hasRequiredPermission" class="permission-wrapper">
    <slot />
  </div>
  <div v-else-if="showFallback" class="permission-fallback">
    <slot name="fallback">
      <div class="no-permission-message">
        <el-icon class="permission-icon"><Lock /></el-icon>
        <span>{{ fallbackMessage }}</span>
      </div>
    </slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Lock } from '@element-plus/icons-vue'
import { usePermissionStore } from '@/stores/permissions'

interface Props {
  // 单个权限检查
  permission?: string
  // 多个权限检查（需要全部拥有）
  permissions?: string[]
  // 多个权限检查（拥有任一即可）
  anyPermissions?: string[]
  // 是否显示无权限时的回退内容
  showFallback?: boolean
  // 自定义无权限提示信息
  fallbackMessage?: string
  // 是否需要管理员身份
  requireAdmin?: boolean
  // 是否需要超级管理员身份
  requireSuperAdmin?: boolean
  // 运动类型ID（用于运动类型访问权限检查）
  sportTypeId?: number
}

const props = withDefaults(defineProps<Props>(), {
  showFallback: false,
  fallbackMessage: '您没有权限访问此功能',
  requireAdmin: false,
  requireSuperAdmin: false
})

const permissionStore = usePermissionStore()

// 权限检查逻辑
const hasRequiredPermission = computed(() => {
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
  
  // 如果没有指定任何权限要求，默认允许访问
  return true
})
</script>

<style scoped>
.permission-wrapper {
  width: 100%;
}

.permission-fallback {
  width: 100%;
}

.no-permission-message {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  color: #909399;
  font-size: 14px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  text-align: center;
}

.permission-icon {
  font-size: 16px;
  color: #c0c4cc;
}

/* 内联样式变体 */
.permission-wrapper.inline {
  display: inline-block;
  width: auto;
}

.permission-fallback.inline {
  display: inline-block;
  width: auto;
}

.permission-fallback.inline .no-permission-message {
  display: inline-flex;
  padding: 4px 8px;
  font-size: 12px;
  background: transparent;
  border: none;
  color: #c0c4cc;
}
</style>