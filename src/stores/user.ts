import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, register as registerApi } from '@/api/auth'
import { usePermissionStore } from './permissions'
import { useAdminStore } from './admin'
import type { User, LoginData, RegisterData } from '@/types/user'


export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  // 初始化本地用户
  const normalizeAvatar = (avatar?: string | null): string | undefined => {
    if (!avatar) return undefined
    const [pathPart, query] = avatar.split('?')
    const filename = pathPart.split('/').pop() || ''
    if (!filename) return undefined
    return `/api/uploads/avatar/${filename}${query ? `?${query}` : ''}`
  }

  try {
    const userJson = localStorage.getItem('user')
    if (userJson) {
      const parsed = JSON.parse(userJson) as User
      if (parsed.avatar) {
        const normalized = normalizeAvatar(parsed.avatar)
        if (normalized) parsed.avatar = normalized
      }
      user.value = parsed
    }
  } catch (error) {
    localStorage.removeItem('user')
  }

  // 计算属性
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // 用户名或昵称
  const displayName = computed(() => {
    if (!user.value) return ''
    return user.value.nickname || user.value.username
  })

  // 是否为管理员
  const isAdmin = computed(() => {
    return user.value?.role === 'admin'
  })

  // 是否为超级管理员（需要结合权限store判断）
  const isSuperAdmin = computed(() => {
    return isAdmin.value && user.value?.role === 'admin'
    // 注意：具体的超级管理员判断逻辑在权限store中
  })

  // 登录方法
  const login = async (credentials: LoginData) => {
    const response = await loginApi(credentials)
    if (!response || !response.access_token) {
      throw new Error('登录失败，服务器无响应')
    }

    token.value = response.access_token
    localStorage.setItem('token', response.access_token)

    if (response.user) {
      // 归一化头像
      if (response.user.avatar) {
        const normalized = normalizeAvatar(response.user.avatar)
        response.user.avatar = normalized || ''
      }
      user.value = response.user
      localStorage.setItem('user', JSON.stringify(response.user))
    } else {
      throw new Error('登录成功但未获取到用户信息')
    }

    // 管理员权限初始化可按需触发，默认不阻塞登录
    if (response.user.role === 'admin') {
      try {
        const permissionStore = usePermissionStore()
        const adminStore = useAdminStore()
        await Promise.all([
          permissionStore.initializePermissions(),
          adminStore.initializeAdminData(),
        ])
      } catch (_) {
        // 忽略权限初始化失败
      }
    }

    return response.user
  }

  // 注册方法
  const register = async (data: RegisterData) => {
    return await registerApi(data)
  }

  // 登出方法
  const logout = async () => {
    // 清除权限数据
    try {
      const permissionStore = usePermissionStore()
      const adminStore = useAdminStore()
      permissionStore.clearPermissions()
      adminStore.resetState()
    } catch (_) {}

    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')

    console.log('登出完成，状态已清除')
  }

  // 更新用户信息
  const updateUserInfo = (userInfo: Partial<User>) => {
    if (user.value) {
      // 更新内存中的用户信息
      user.value = { ...user.value, ...userInfo }

      // 归一化头像 URL，统一使用 /api/uploads/avatar/{filename}
      if (userInfo.avatar) {
        const normalized = normalizeAvatar(userInfo.avatar)
        if (normalized) {
          user.value.avatar = normalized
        }
      }

      // 保存到 localStorage（头像也已归一化）
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    displayName,
    isAdmin,
    isSuperAdmin,
    login,
    logout,
    register,
    updateUserInfo
  }
})

export default useUserStore