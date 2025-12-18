// 权限指令
import type { Directive, DirectiveBinding } from 'vue'
import { usePermissionStore } from '@/stores/permissions'

interface PermissionBinding {
  // 单个权限
  permission?: string
  // 多个权限（需要全部拥有）
  permissions?: string[]
  // 多个权限（拥有任一即可）
  anyPermissions?: string[]
  // 是否需要管理员身份
  requireAdmin?: boolean
  // 是否需要超级管理员身份
  requireSuperAdmin?: boolean
  // 运动类型ID
  sportTypeId?: number
  // 权限检查失败时的处理方式：'hide' | 'disable' | 'class'
  action?: 'hide' | 'disable' | 'class'
  // 权限检查失败时添加的CSS类名
  failedClass?: string
}

// 权限检查函数
function checkPermission(binding: PermissionBinding): boolean {
  const permissionStore = usePermissionStore()
  
  // 超级管理员检查
  if (binding.requireSuperAdmin) {
    return permissionStore.isSuperAdmin
  }
  
  // 管理员身份检查
  if (binding.requireAdmin) {
    if (!permissionStore.isAdmin) {
      return false
    }
  }
  
  // 运动类型访问权限检查
  if (binding.sportTypeId !== undefined) {
    if (!permissionStore.canAccessSportType(binding.sportTypeId)) {
      return false
    }
  }
  
  // 单个权限检查
  if (binding.permission) {
    return permissionStore.hasPermission(binding.permission)
  }
  
  // 多个权限检查（需要全部拥有）
  if (binding.permissions && binding.permissions.length > 0) {
    return permissionStore.hasAllPermissions(binding.permissions)
  }
  
  // 多个权限检查（拥有任一即可）
  if (binding.anyPermissions && binding.anyPermissions.length > 0) {
    return permissionStore.hasAnyPermission(binding.anyPermissions)
  }
  
  // 如果只要求管理员身份且已通过检查，则允许访问
  if (binding.requireAdmin && permissionStore.isAdmin) {
    return true
  }
  
  // 默认拒绝访问
  return false
}

// 处理权限检查结果
function handlePermissionResult(el: HTMLElement, hasPermission: boolean, binding: PermissionBinding) {
  const action = binding.action || 'hide'
  const failedClass = binding.failedClass || 'permission-denied'
  
  if (hasPermission) {
    // 有权限时恢复元素状态
    el.style.display = ''
    el.removeAttribute('disabled')
    el.classList.remove(failedClass)
  } else {
    // 无权限时根据action处理
    switch (action) {
      case 'hide':
        el.style.display = 'none'
        break
      case 'disable':
        if (el.tagName === 'BUTTON' || el.tagName === 'INPUT') {
          el.setAttribute('disabled', 'true')
        } else {
          el.style.pointerEvents = 'none'
          el.style.opacity = '0.5'
        }
        break
      case 'class':
        el.classList.add(failedClass)
        break
    }
  }
}

// v-permission 指令
export const vPermission: Directive<HTMLElement, string | PermissionBinding> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'string') {
      // 简单字符串形式：v-permission="'permission.code'"
      permissionBinding = { permission: binding.value }
    } else {
      // 对象形式：v-permission="{ permission: 'code', action: 'disable' }"
      permissionBinding = binding.value || {}
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<string | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'string') {
      permissionBinding = { permission: binding.value }
    } else {
      permissionBinding = binding.value || {}
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  }
}

// v-admin 指令（管理员权限检查的简化版本）
export const vAdmin: Directive<HTMLElement, boolean | PermissionBinding> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<boolean | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'boolean') {
      // 简单布尔形式：v-admin="true" 或 v-admin
      permissionBinding = { requireAdmin: binding.value !== false }
    } else {
      // 对象形式
      permissionBinding = { requireAdmin: true, ...binding.value }
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<boolean | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'boolean') {
      permissionBinding = { requireAdmin: binding.value !== false }
    } else {
      permissionBinding = { requireAdmin: true, ...binding.value }
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  }
}

// v-super-admin 指令（超级管理员权限检查）
export const vSuperAdmin: Directive<HTMLElement, boolean | PermissionBinding> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<boolean | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'boolean') {
      permissionBinding = { requireSuperAdmin: binding.value !== false }
    } else {
      permissionBinding = { requireSuperAdmin: true, ...binding.value }
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<boolean | PermissionBinding>) {
    let permissionBinding: PermissionBinding
    
    if (typeof binding.value === 'boolean') {
      permissionBinding = { requireSuperAdmin: binding.value !== false }
    } else {
      permissionBinding = { requireSuperAdmin: true, ...binding.value }
    }
    
    const hasPermission = checkPermission(permissionBinding)
    handlePermissionResult(el, hasPermission, permissionBinding)
  }
}

// 导出所有指令
export default {
  permission: vPermission,
  admin: vAdmin,
  superAdmin: vSuperAdmin
}