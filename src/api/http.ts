import axios, { AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const baseURL = import.meta.env.VITE_API_BASE_URL || ''

const instance = axios.create({
  baseURL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

let handling401 = false

instance.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    const token = userStore.token

    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }

    return config
  },
  (error) => Promise.reject(error)
)

const extractErrorMessage = (data: any, fallback = '请求失败') => {
  if (!data || typeof data !== 'object') return fallback
  let msg = data.message || fallback
  if (data.error?.details) msg += ` (${data.error.details})`
  if (data.details) {
    if (Array.isArray(data.details)) {
      const detailMessages = data.details.map((detail: any) =>
        `${detail.property}: ${detail.message || JSON.stringify(detail)}`
      )
      msg += ` (${detailMessages.join('; ')})`
    } else if (typeof data.details === 'string') {
      msg += ` (${data.details})`
    }
  }
  return msg
}

instance.interceptors.response.use(
  (response) => {
    if (response.data && typeof response.data === 'object') {
      if (response.data.success !== undefined) {
        if (response.data.success === false) {
          const errorMessage = extractErrorMessage(response.data, '请求失败')
          return Promise.reject(new Error(errorMessage))
        }

        if (response.data.success === true) {
          const result = response.data.data !== undefined ? response.data.data : response.data
          return result
        }
      }
    }
    return response.data
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response

      if (status === 401) {
        const userStore = useUserStore()
        userStore.logout()

        if (!handling401 && !window.location.pathname.includes('/login')) {
          handling401 = true
          ElMessage.error('登录已过期，请重新登录')
          setTimeout(() => {
            window.location.href = '/login'
            handling401 = false
          }, 1200)
        }
      }

      const errorMessage = extractErrorMessage(data)
      ElMessage.error(errorMessage)
      return Promise.reject(new Error(errorMessage))
    }

    const message = error.message || '网络错误'
    ElMessage.error(message)
    return Promise.reject(new Error(message))
  }
)

// GET请求
export const get = <T>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> => {
  return instance.get(url, { params, ...config })
}

// POST请求
export const post = <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return instance.post(url, data, config)
}

// PUT请求
export const put = <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return instance.put(url, data, config)
}

// PATCH请求
export const patch = <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return instance.patch(url, data, config)
}

// DELETE请求
export const del = <T>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  return instance.delete(url, config)
}

export default instance