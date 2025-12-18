import { get, post } from './http'
import { LoginData, RegisterData, User, LoginResponse, RegisterResponse } from '@/types/user'

/**
 * 用户登录
 */
export const login = async (data: LoginData): Promise<LoginResponse> => {
  try {
    let response: any
    try {
      response = await post<any>('/api/auth/login', data)
    } catch (err: any) {
      // 兼容旧路径：若 404/405 则退回 /auth/login
      const status = err?.response?.status
      if (status === 404 || status === 405) {
        response = await post<any>('/auth/login', data)
      } else {
        throw err
      }
    }

    // 确保响应包含必要的数据
    if (!response || !response.access_token) {
      throw new Error('登录响应格式不正确')
    }

    console.log('登录成功，响应数据：', response)
    return response
  } catch (error) {
    console.error('登录失败：', error)
    throw error
  }
}

/**
 * 用户注册
 */
export const register = (data: RegisterData): Promise<RegisterResponse> => {
  // 创建一个新对象，去除password_confirm字段
  const { password_confirm, ...registerData } = data;

  console.log('发送注册请求，数据:', registerData);

  // 后端直接返回用户对象，需要转换成前端期望的格式
  return post<any>('/api/auth/register', registerData)
    .then(response => {
      console.log('注册成功，响应:', response);
      return {
        success: true,
        message: '注册成功'
      }
    })
    .catch(error => {
      console.error('注册失败，错误:', error);
      throw error;
    });
}

/**
 * 用户登出
 */
export const logout = (): Promise<void> => {
  return post<void>('/api/auth/logout')
}

/**
 * 获取用户个人资料
 */
export const fetchUserProfile = (): Promise<User> => {
  return get<User>('/api/auth/profile')
}

/**
 * 更新用户个人资料
 */
export const updateUserProfile = (data: any): Promise<User> => {
  // 如果包含文件，则需要使用FormData
  if (data.avatar instanceof File) {
    const formData = new FormData()

    if (data.nickname) {
      formData.append('nickname', data.nickname)
    }

    if (data.email) {
      formData.append('email', data.email)
    }

    if (data.avatar) {
      formData.append('avatar', data.avatar)
    }

    return post<User>('/api/auth/profile', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }

  // 普通JSON请求
  return post<User>('/api/auth/profile', data)
}

/**
 * 修改密码
 */
export const changePassword = (
  currentPassword: string,
  newPassword: string,
  newPasswordConfirm: string
): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>('/api/auth/change-password', {
    currentPassword,
    newPassword,
    newPasswordConfirm
  })
}

/**
 * 检查用户名是否可用
 */
export const checkUsernameAvailability = (username: string): Promise<{ available: boolean }> => {
  return get<{ available: boolean }>('/users/check-username', { username })
}